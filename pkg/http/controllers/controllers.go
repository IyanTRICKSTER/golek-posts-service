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

	router.Use(middleware.HandleCORS())

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
	//r.Use(middleware.ValidateRequestHeaderMiddleware)
	r.GET("/list", middleware.ValidateRequestHeaderMiddleware, postHandler.Fetch)
	r.GET("/:id", middleware.ValidateRequestHeaderMiddleware, postHandler.FetchByID)
	r.GET("/s/:keyword", middleware.ValidateRequestHeaderMiddleware, postHandler.Search)
	r.POST("/", middleware.ValidateRequestHeaderMiddleware, postHandler.Create)
	r.PUT("/:id", middleware.ValidateRequestHeaderMiddleware, postHandler.Update)
	r.DELETE("/:id", middleware.ValidateRequestHeaderMiddleware, postHandler.Delete)
	r.GET("/validate/:user_id", middleware.ValidateRequestHeaderMiddleware, postHandler.ReqValidateOwner)
	r.POST("/validate/", middleware.ValidateRequestHeaderMiddleware, postHandler.ValidateOwner)

}
