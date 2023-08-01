package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"golek_posts_service/pkg/contracts"
	"golek_posts_service/pkg/contracts/status"
	"golek_posts_service/pkg/http/middleware"
	"golek_posts_service/pkg/http/requests"
	"golek_posts_service/pkg/models"
	"golek_posts_service/pkg/utils"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostService struct {
	StorageRepository   contracts.ICloudStorageRepo
	PostRepository      contracts.PostRepositoryContract
	QrCodeRepository    contracts.QrCodeRepository
	MessageQueueService contracts.MessageQueue
}

func NewPostService(postRepository *contracts.PostRepositoryContract,
	qrcodeRepository *contracts.QrCodeRepository, storageRepository *contracts.ICloudStorageRepo, mqService *contracts.MessageQueue) contracts.PostServiceContract {

	return &PostService{
		PostRepository:      *postRepository,
		QrCodeRepository:    *qrcodeRepository,
		StorageRepository:   *storageRepository,
		MessageQueueService: *mqService,
	}
}

func (p PostService) Search(ctx context.Context, keyword string, pagination models.Pagination) ([]models.Post, error) {
	limit, skip := pagination.GetPagination()

	posts, err := p.PostRepository.Search(ctx, keyword, limit, skip)
	if err != nil {
		return []models.Post{}, err
	}

	return posts, nil
}

func (p PostService) RequestValidateOwner(ctx context.Context, postID string) (qrCode string, opStatus status.PostOperationStatus, err error) {

	post, err := p.PostRepository.FindById(ctx, postID)
	if err != nil {
		return "", status.PostRequestValidationFailed, err
	}

	if post.IsReturned {
		return "", status.PostAlreadyReturned, nil
	}

	authenticatedReq := ctx.Value("authenticatedRequest").(*middleware.AuthenticatedRequest)

	//Checking authorization
	opStatus, err = ProtectResource(
		contracts.Resource{
			Alias: "vp",
			Name:  "Validate",
		},
		authenticatedReq,
		post,
		func(isOwner bool) (opStatus status.PostOperationStatus, err error) {
			//if !isOwner {
			//	return status.OperationForbidden, errors.New(
			//		fmt.Sprintf("User x is not the owner of Model y"))
			//}
			return status.OperationAllowed, nil
		},
	)

	if err != nil {
		return "", opStatus, err
	}

	//1. hash post id with hash secret
	hash := utils.Hash(os.Getenv("VALIDATION_SECRET") + postID)

	//2. create json string contains post id and hashed post id
	data := fmt.Sprintf(`{"post_id": "%s", "hash": "%s"}`, postID, hash)

	//3. generate qrcode with data from step 2
	qrcodeUrl, err := p.QrCodeRepository.Generate(data)
	if err != nil {
		return "", status.PostRequestValidationFailed, err
	}

	//4. Assign hashed data to post
	post.ConfirmationKey = hash

	//5. Update post
	_, opStatus, err = p.PostRepository.Update(ctx, postID, post)
	if err != nil || opStatus == status.PostUpdatedStatusFailed {
		return "", status.PostRequestValidationFailed, err
	}

	return qrcodeUrl, status.PostRequestValidationSuccess, nil
}

func (p PostService) ValidateOwner(ctx context.Context, request requests.ValidateItemOwnerRequest) (opStatus status.PostOperationStatus, err error) {

	post, err := p.PostRepository.FindById(ctx, request.PostID)
	if err != nil {
		return status.PostValidateOwnerFailed, err
	}

	if post.IsReturned {
		return status.PostAlreadyReturned, nil
	}

	authenticatedReq := ctx.Value("authenticatedRequest").(*middleware.AuthenticatedRequest)

	//Checking authorization
	opStatus, err = ProtectResource(
		contracts.Resource{
			Alias: "vp",
			Name:  "Validate",
		},
		authenticatedReq,
		post,
		func(isOwner bool) (opStatus status.PostOperationStatus, err error) {
			//if !isOwner {
			//	return status.OperationForbidden, errors.New(
			//		fmt.Sprintf("User x is not the owner of Model y"))
			//}
			return status.OperationAllowed, nil
		},
	)

	if err != nil {
		return opStatus, err
	}

	if !utils.HashCompare(request.Hash, os.Getenv("VALIDATION_SECRET")+request.PostID) && post.ConfirmationKey != request.Hash {
		return status.PostValidateOwnerFailed, errors.New("unmatched given hash")
	}

	userID, _ := strconv.Atoi(authenticatedReq.UserID)
	post.ReturnedTo = int64(userID)
	post.IsReturned = true

	_, opStatus, err2 := p.PostRepository.Update(ctx, request.PostID, post)
	if err2 != nil || opStatus == status.PostUpdatedStatusFailed {
		return status.PostValidateOwnerFailed, err
	}

	return status.PostValidateOwnerSuccess, nil

}

func (p PostService) Fetch(ctx context.Context, pagination models.Pagination, filter map[string]any) ([]models.Post, error) {
	limit, skip := pagination.GetPagination()

	filters := map[string]any{"deleted_at": nil}
	maps.Copy(filters, filter)

	posts, err := p.PostRepository.Fetch(ctx, true, limit, skip, filters)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (p PostService) FindById(ctx context.Context, postID string) (models.Post, error) {
	post, err := p.PostRepository.FindById(ctx, postID)
	if err != nil {
		return models.Post{}, err
	}
	return post, nil
}

func (p PostService) Create(ctx context.Context, request requests.CreatePostRequest) (models.Post, status.PostOperationStatus, error) {

	authenticatedReq := ctx.Value("authenticatedRequest").(*middleware.AuthenticatedRequest)

	//Checking authorization
	opStatus, err := ProtectResource(
		contracts.Resource{
			Alias: "c",
			Name:  "Create",
		},
		authenticatedReq,
		models.Post{UserID: 10},
		func(isOwner bool) (opStatus status.PostOperationStatus, err error) {
			return status.OperationAllowed, nil
		},
	)

	postCharacteristics := make([]models.Characteristic, 0)

	for _, d := range request.Characteristics {
		postCharacteristics = append(postCharacteristics,
			models.Characteristic{Title: d.Title},
		)
	}

	userID, err := strconv.Atoi(authenticatedReq.UserID)
	if err != nil {
		return models.Post{}, status.PostCreatedStatusFailed, err
	}

	timeNow := time.Now()

	//Upload an image to Storage
	fileUrl, err := p.StorageRepository.UploadFile(ctx, request.Image, request.Title+UnixMilliToStr())
	if err != nil {
		return models.Post{}, status.PostCreatedStatusFailed, err
	}

	newPost := models.Post{
		ID:              primitive.NewObjectID(),
		UserID:          int64(userID),
		ReturnedTo:      0,
		IsReturned:      false,
		Title:           request.Title,
		ImageURL:        fileUrl,
		ImageKey:        "",
		ConfirmationKey: "",
		Place:           request.Place,
		Description:     request.Description,
		Characteristics: postCharacteristics,
		UpdatedAt:       &timeNow,
		CreatedAt:       &timeNow,
		DeletedAt:       nil,
		User: models.UserInfo{
			Username:  authenticatedReq.Username,
			UserMajor: authenticatedReq.UserMajor,
		},
	}

	createdPost, opStatus, err := p.PostRepository.Create(ctx, newPost)
	if err != nil || opStatus == status.PostCreatedStatusFailed {
		return models.Post{}, status.PostCreatedStatusFailed, err
	}

	//Notify all User
	go func() {
		payload, err := json.Marshal(contracts.MessagePayload{
			UserID:   int64(userID),
			Title:    "Telah ditemukan " + createdPost.Title,
			Body:     createdPost.Characteristics[0].Title,
			ImageUrl: createdPost.ImageURL,
		})
		if err != nil {
			log.Printf("Bookmark Service: Create >> %v", err)
		}

		err = p.MessageQueueService.Publish(payload)
		if err != nil {
			log.Printf("Bookmark Service: Create >> Error Broadcast Message: %v", err)
		}
	}()

	return createdPost, status.PostCreatedStatusSuccess, nil
}

func (p PostService) Update(ctx context.Context, postID string, request requests.UpdatePostRequest) (models.Post, status.PostOperationStatus, error) {

	//Load particular post to be checked
	post, err := p.PostRepository.FindById(ctx, postID)
	if err != nil {
		return models.Post{}, status.PostUpdatedStatusFailed, err
	}

	//Checking authorization
	opStatus, err := ProtectResource(
		contracts.Resource{
			Alias: "u",
			Name:  "Update",
		},
		ctx.Value("authenticatedRequest").(*middleware.AuthenticatedRequest),
		post,
		func(isOwner bool) (opStatus status.PostOperationStatus, err error) {
			if !isOwner {
				return status.OperationForbidden, errors.New(
					fmt.Sprintf("User x is not the owner of Model y"))
			}
			return status.OperationAllowed, nil
		},
	)

	if err != nil {
		return models.Post{}, opStatus, err
	}

	timeNow := time.Now()

	//If authorized then,
	//Delete old image and upload a new one (if image exists in update request)
	if request.Image != nil {
		err := p.StorageRepository.DeleteFile(ctx, post.ImageURL)
		if err != nil {
			return models.Post{}, status.PostUpdatedStatusFailed, err
		}

		post.ImageURL, err = p.StorageRepository.UploadFile(ctx, request.Image, request.Title+UnixMilliToStr())
		if err != nil {
			return models.Post{}, status.PostUpdatedStatusFailed, err
		}
	}

	//Update post data
	post.Title = request.Title
	post.Place = request.Place
	post.Description = request.Description
	post.UpdatedAt = &timeNow

	//I indicate changes in characteristic by comparing len of the old and the new input
	if len(post.Characteristics) != len(request.Characteristics) {
		postCharacteristics := make([]models.Characteristic, 0)

		for _, d := range request.Characteristics {
			postCharacteristics = append(postCharacteristics,
				models.Characteristic{Title: d.Title},
			)
		}

		post.Characteristics = postCharacteristics
	}

	updatePost, opStatus, err := p.PostRepository.Update(ctx, postID, post)
	if err != nil || opStatus == status.PostUpdatedStatusFailed {
		return models.Post{}, status.PostUpdatedStatusFailed, err
	}
	return updatePost, status.PostUpdatedStatusSuccess, nil
}

func (p PostService) Delete(ctx context.Context, postID string) (status.PostOperationStatus, error) {

	post, err := p.PostRepository.FindById(ctx, postID)
	if err != nil {
		return status.PostDeletedStatusFailed, err
	}

	opStatus, err := ProtectResource(
		contracts.Resource{
			Alias: "d",
			Name:  "Delete",
		},
		ctx.Value("authenticatedRequest").(*middleware.AuthenticatedRequest),
		post,
		func(isOwner bool) (opStatus status.PostOperationStatus, err error) {
			if !isOwner {
				return status.OperationForbidden, errors.New(
					fmt.Sprintf("User x is not the owner of Model y"))
			}
			return status.OperationAllowed, nil
		},
	)

	if err != nil {
		return opStatus, err
	}

	err = p.StorageRepository.DeleteFile(ctx, post.ImageURL)
	if err != nil {
		return status.PostDeletedStatusFailed, err
	}

	opStatus, err = p.PostRepository.Delete(ctx, postID)
	if err != nil || opStatus == status.PostDeletedStatusFailed {
		return status.PostDeletedStatusFailed, err
	}
	return status.PostDeletedStatusSuccess, nil
}

// ProtectResource Test
func ProtectResource(resource contracts.Resource, authenticated *middleware.AuthenticatedRequest, model models.Post, callback func(isOwner bool) (opStatus status.PostOperationStatus, err error)) (status.PostOperationStatus, error) {

	//log.Println("Checking User Permissions")

	//Check User Authorization
	if !strings.Contains(authenticated.Permissions, resource.Alias) {
		return status.OperationUnauthorized, errors.New("You doesn't have any permission to access " + resource.Name + " resource")
	}

	log.Printf("User id %v is authorized", authenticated.UserID)

	//Check Model's owner
	isOwner := authenticated.UserID == strconv.FormatInt(model.UserID, 10)

	//log.Printf("Is Owner Checking >> User id comparison %v by %v is %v", authenticated.UserID, model.UserID, isOwner)

	//Run the callback
	opStatus, err := callback(isOwner)
	if err != nil {
		return opStatus, err
	}

	return opStatus, nil
}

func (p PostService) replaceVideoUrl(url string) string {
	var regex, _ = regexp.Compile(os.Getenv("AWS_S3_DEFAULT_OBJECT_URL"))
	return regex.ReplaceAllString(url, os.Getenv("AWS_CLOUDFRONT_URL"))
}

func UnixMilliToStr() string {
	return strconv.Itoa(int(time.Now().UnixMilli()))
}
