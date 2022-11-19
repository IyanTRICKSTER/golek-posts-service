package services

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
)

type PostService struct {
	StorageRepository contracts.StorageRepository
	PostRepository    contracts.PostRepositoryContract
	QrCodeRepository  contracts.QrCodeRepository
}

func NewPostService(postRepository *contracts.PostRepositoryContract,
	qrcodeRepository *contracts.QrCodeRepository, storageRepository *contracts.StorageRepository) contracts.PostServiceContract {

	return &PostService{
		PostRepository:    *postRepository,
		QrCodeRepository:  *qrcodeRepository,
		StorageRepository: *storageRepository,
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

	//1. hash post id with hash secret
	hash := utils.Hash("mysecret" + postID)

	//2. create json string contains post id and hashed post id
	data := fmt.Sprintf("{\"post_id\": \"%s\", \"hash\": \"%s\"}", postID, hash)

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

func (p PostService) ValidateOwner(ctx context.Context, request requests.ValidateItemOwnerRequest) (status.PostOperationStatus, error) {

	post, err := p.PostRepository.FindById(ctx, request.PostID)
	if err != nil {
		return status.PostValidateOwnerFailed, err
	}

	if post.IsReturned {
		return status.PostAlreadyReturned, nil
	}

	if !utils.HashCompare(request.Hash, "mysecret"+request.PostID) && post.ConfirmationKey != request.Hash {
		return status.PostValidateOwnerFailed, errors.New("unmatched given hash")
	}

	post.ReturnedTo = request.OwnerID
	post.IsReturned = true

	_, opStatus, err2 := p.PostRepository.Update(ctx, request.PostID, post)
	if err2 != nil || opStatus == status.PostUpdatedStatusFailed {
		return status.PostValidateOwnerFailed, err
	}

	return status.PostValidateOwnerSuccess, nil

}

func (p PostService) Fetch(ctx context.Context, pagination models.Pagination) ([]models.Post, error) {
	limit, skip := pagination.GetPagination()
	posts, err := p.PostRepository.Fetch(ctx, true, limit, skip)
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

	postCharacteristics := make([]models.Characteristic, 0)

	for _, d := range request.Characteristics {
		postCharacteristics = append(postCharacteristics,
			models.Characteristic{Title: d.Title},
		)
	}

	//Upload an image to Storage
	s3Response, err := p.StorageRepository.UploadFile(request.Image, "")
	if err != nil {
		return models.Post{}, 0, err
	}

	timeNow := time.Now()
	newPost := models.Post{
		ID:              primitive.NewObjectID(),
		UserID:          request.UserID,
		ReturnedTo:      0,
		IsReturned:      false,
		Title:           request.Title,
		ImageURL:        p.replaceVideoUrl(s3Response.Filepath),
		ImageKey:        s3Response.Key,
		ConfirmationKey: "",
		Place:           request.Place,
		Description:     request.Description,
		Characteristics: postCharacteristics,
		UpdatedAt:       &timeNow,
		CreatedAt:       &timeNow,
		DeletedAt:       nil,
	}

	createdPost, opStatus, err := p.PostRepository.Create(ctx, newPost)
	if err != nil || opStatus == status.PostCreatedStatusFailed {
		return models.Post{}, status.PostCreatedStatusFailed, err
	}

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

	//If authorized then,
	//Delete old image and upload a new one (if image exists in update request)
	if request.Image != nil {
		err := p.StorageRepository.DeleteObject(&post.ImageKey)
		if err != nil {
			return models.Post{}, status.PostUpdatedStatusFailed, err
		}

		s3Response, err := p.StorageRepository.UploadFile(request.Image, "")
		if err != nil {
			return models.Post{}, status.PostUpdatedStatusFailed, err
		}

		post.ImageURL = p.replaceVideoUrl(s3Response.Filepath)
		post.ImageKey = s3Response.Key
	}

	//Update post data
	timeNow := time.Now()
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

	err = p.StorageRepository.DeleteObject(&post.ImageKey)
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

	log.Println("Checking User Permissions")

	//Check User Authorization
	if !strings.Contains(resource.Alias, authenticated.Permissions) {
		return status.OperationUnauthorized, errors.New("User x doesn't have any permission to access " + resource.Name + " resource")
	}

	//Check Model's owner
	isOwner := authenticated.UserID == strconv.Itoa(int(model.UserID))

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
