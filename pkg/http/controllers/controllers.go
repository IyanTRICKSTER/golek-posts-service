package controllers

import (
	"github.com/gin-gonic/gin"
	"golek_posts_service/pkg/contracts"
	"golek_posts_service/pkg/http/middleware"
	"golek_posts_service/pkg/http/responses"
	"net/http"
)

func SetupHandler(router *gin.Engine, postService *contracts.PostServiceContract) {

	postHandler := PostHandler{*postService}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, responses.HttpResponse{
			StatusCode: http.StatusNotFound,
			Message:    "PAGE NOT FOUND",
			Data:       nil,
		})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, responses.HttpResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "METHOD NOT ALLOWED",
			Data:       nil,
		})
	})

	r := router.Group("/api/posts/")
	r.Use(middleware.ValidateRequestHeaderMiddleware)
	r.GET("/list", postHandler.Fetch)
	r.GET("/:id", postHandler.FetchByID)
	r.GET("/s/:keyword", postHandler.Search)
	r.POST("/", postHandler.Create)
	r.PUT("/:id", postHandler.Update)
	r.DELETE("/:id", postHandler.Delete)
	r.GET("/validate/:user_id", postHandler.ReqValidateOwner)
	r.POST("/validate/", postHandler.ValidateOwner)

}
