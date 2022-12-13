package controllers

import (
	"context"
	"golek_posts_service/pkg/contracts"
	"golek_posts_service/pkg/contracts/status"
	"golek_posts_service/pkg/http/requests"
	"golek_posts_service/pkg/http/responses"
	"golek_posts_service/pkg/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostHandler struct {
	PostService contracts.PostServiceContract
}

func (h *PostHandler) Fetch(c *gin.Context) {

	//Get page number
	page, ok := c.GetQuery("page")
	if page == "" || !ok {
		page = "1"
	}

	//get limit number
	limit, ok := c.GetQuery("limit")
	if limit == "" || !ok {
		limit = "10"
	}

	userID, ok := c.GetQuery("user-id")
	if userID == "" || !ok {
		userID = "0"
	}

	isReturned := false
	returned, ok := c.GetQuery("returned")
	if ok && returned == "1" {
		isReturned = true
	}

	//convert page parameter to int
	qPage, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Parsing Page Parameter " + err.Error(),
		})
		return
	}

	//convert limit parameter to int
	qLimit, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Parsing Limit Parameter " + err.Error(),
		})
		return
	}

	//convert user id parameter to int
	userIDint, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Parsing UserID Parameter " + err.Error(),
		})
		return
	}

	paginate := models.Pagination{
		Page:    qPage,
		PerPage: qLimit,
	}

	filter := map[string]any{"is_returned": isReturned}
	if userIDint != 0 {
		filter["user_id"] = userIDint
	}
	posts, err := h.PostService.Fetch(context.TODO(), paginate, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "PostService Fetch " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.HttpPaginationResponse{
		HttpResponse: responses.HttpResponse{
			Data:       posts,
			StatusCode: 200,
		},
		PerPage: paginate.PerPage,
		Page:    paginate.Page,
	})
	return
}

func (h *PostHandler) FetchByID(c *gin.Context) {
	post, err := h.PostService.FindById(context.Background(), c.Param("id"))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.HttpResponse{
		StatusCode: http.StatusOK,
		Message:    "data exists",
		Data:       post,
	})
	return
}

func (h *PostHandler) Search(c *gin.Context) {

	page, ok := c.GetQuery("page")
	if page == "" || !ok {
		page = "1"
	}

	limit, ok := c.GetQuery("limit")
	if page == "" || !ok {
		page = "1"
	}

	qPage, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Parsing Page Parameter " + err.Error(),
		})
		return
	}

	qLimit, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Parsing Limit Parameter " + err.Error(),
		})
		return
	}

	paginate := models.Pagination{
		Page:    qPage,
		PerPage: qLimit,
	}

	posts, err := h.PostService.Search(context.TODO(), c.Param("keyword"), paginate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.HttpPaginationResponse{
		PerPage: qLimit,
		Page:    qPage,
		HttpResponse: responses.HttpResponse{
			StatusCode: http.StatusOK,
			Message:    "",
			Data:       posts,
		},
	})
	return
}

func (h *PostHandler) Create(c *gin.Context) {

	var createReq requests.CreatePostRequest

	val, _ := c.Get("authenticatedRequest")
	authContext := context.WithValue(context.Background(), "authenticatedRequest", val)

	err := c.ShouldBind(&createReq)
	if err != nil {
		log.Printf("Request Binding Error: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Binding error " + err.Error(),
		})
		return
	}

	createdPost, opStatus, err := h.PostService.Create(authContext, createReq)
	if err != nil || opStatus == status.PostCreatedStatusFailed {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "PostService Create " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post created successfully",
		"data":    createdPost,
	})
	return
}

func (h *PostHandler) Update(c *gin.Context) {

	var updateReq requests.UpdatePostRequest

	val, _ := c.Get("authenticatedRequest")
	authContext := context.WithValue(context.Background(), "authenticatedRequest", val)

	err := c.ShouldBind(&updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Binding error " + err.Error(),
		})
		return
	}

	updatedPost, opStatus, err := h.PostService.Update(authContext, c.Param("id"), updateReq)

	if opStatus == status.OperationUnauthorized {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "PostService Update " + err.Error(),
		})
		return
	}

	if opStatus == status.OperationForbidden {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "PostService Update " + err.Error(),
		})
		return
	}

	if err != nil && opStatus == status.PostUpdatedStatusFailed {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "PostService Update " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"data":    updatedPost,
	})
	return
}

func (h *PostHandler) Delete(c *gin.Context) {

	val, _ := c.Get("authenticatedRequest")
	authContext := context.WithValue(context.Background(), "authenticatedRequest", val)

	opStatus, err := h.PostService.Delete(authContext, c.Param("id"))

	if opStatus == status.OperationUnauthorized {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "PostService Delete " + err.Error(),
		})
		return
	}

	if opStatus == status.OperationForbidden {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "PostService Delete " + err.Error(),
		})
		return
	}

	if err != nil || opStatus == status.PostDeletedStatusFailed {

		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, responses.HttpResponse{
		StatusCode: http.StatusOK,
		Message:    "data deleted successfully",
		Data:       nil,
	})

	return
}

func (h *PostHandler) ReqValidateOwner(c *gin.Context) {

	postID := c.Param("user_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, responses.HttpErrorResponse{
			StatusCode: http.StatusBadRequest,
			Error:      "Query parameter is not valid",
			Data:       nil,
		})
		return
	}

	val, _ := c.Get("authenticatedRequest")
	authContext := context.WithValue(context.Background(), "authenticatedRequest", val)

	qrCodeUrl, opStatus, err := h.PostService.RequestValidateOwner(authContext, postID)
	if opStatus == status.OperationUnauthorized {
		c.JSON(http.StatusUnauthorized, responses.HttpErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Error:      "Request Validate Owner " + err.Error(),
		})
		return
	}
	if opStatus == status.OperationForbidden {
		c.JSON(http.StatusForbidden, responses.HttpErrorResponse{
			StatusCode: http.StatusForbidden,
			Error:      "Request Validate Owner " + err.Error(),
		})
		return
	}
	if err != nil {
		if opStatus == status.PostRequestValidationFailed {
			c.JSON(http.StatusInternalServerError, responses.HttpErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Error:      "Request Validate Owner " + err.Error(),
			})
			return
		}
	}
	if opStatus == status.PostAlreadyReturned {
		c.JSON(http.StatusOK, responses.HttpResponse{
			StatusCode: http.StatusOK,
			Message:    "post already returned",
			Data:       nil,
		})
		return
	}
	c.JSON(http.StatusOK, responses.HttpResponse{
		StatusCode: http.StatusOK,
		Data:       gin.H{
			"qr_code_url": qrCodeUrl,
		},
	})
	return
}

func (h *PostHandler) ValidateOwner(c *gin.Context) {

	var validatePostReq requests.ValidateItemOwnerRequest

	err := c.ShouldBind(&validatePostReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.HttpErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Error:      err.Error(),
		})
		return
	}

	val, _ := c.Get("authenticatedRequest")
	authContext := context.WithValue(context.Background(), "authenticatedRequest", val)

	opStatus, err := h.PostService.ValidateOwner(authContext, validatePostReq)
	if err != nil {
		if opStatus == status.PostValidateOwnerFailed {
			c.JSON(http.StatusInternalServerError, responses.HttpErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Error:      err.Error(),
			})
			return
		}
		if opStatus == status.OperationUnauthorized {
			c.JSON(http.StatusUnauthorized, responses.HttpErrorResponse{
				StatusCode: http.StatusUnauthorized,
				Error:      "Validate Owner " + err.Error(),
			})
			return
		}
		if opStatus == status.OperationForbidden {
			c.JSON(http.StatusForbidden, responses.HttpErrorResponse{
				StatusCode: http.StatusForbidden,
				Error:      "Validate Owner " + err.Error(),
			})
			return
		}
	}

	if opStatus == status.PostAlreadyReturned {
		c.JSON(http.StatusOK, responses.HttpResponse{
			StatusCode: http.StatusOK,
			Message:    "post already mark as returned",
		})
	}

	if opStatus == status.PostValidateOwnerSuccess {
		c.JSON(http.StatusOK, responses.HttpResponse{
			StatusCode: http.StatusOK,
			Message:    "marking post as returned success",
		})
	}
}
