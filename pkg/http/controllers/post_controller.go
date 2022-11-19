package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golek_posts_service/pkg/contracts"
	"golek_posts_service/pkg/contracts/status"
	"golek_posts_service/pkg/http/requests"
	"golek_posts_service/pkg/http/responses"
	"golek_posts_service/pkg/models"
	"net/http"
	"strconv"
)

type PostHandler struct {
	PostService contracts.PostServiceContract
}

func (h *PostHandler) Fetch(c *gin.Context) {

	page, ok := c.GetQuery("page")
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

	paginate := models.Pagination{
		Page:    qPage,
		PerPage: 25,
	}

	posts, err := h.PostService.Fetch(context.TODO(), paginate)
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

	qPage, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Parsing Page Parameter " + err.Error(),
		})
		return
	}

	paginate := models.Pagination{
		Page:    qPage,
		PerPage: 25,
	}

	posts, err := h.PostService.Search(context.TODO(), c.Param("keyword"), paginate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.HttpPaginationResponse{
		PerPage: 25,
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

	err := c.ShouldBind(&createReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Binding error " + err.Error(),
		})
		return
	}

	createdPost, opStatus, err := h.PostService.Create(context.TODO(), createReq)
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
