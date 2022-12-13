package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthenticatedRequest struct {
	UserID      string
	Role        string
	Permissions string
	Username    string
	UserMajor   string
}

var authenticated *AuthenticatedRequest

func ValidateRequestHeaderMiddleware(c *gin.Context) {

	userPermission := c.Request.Header.Get("X-User-Permission")
	userRole := c.Request.Header.Get("X-User-Role")
	userId := c.Request.Header.Get("X-User-Id")
	username := c.Request.Header.Get("X-User-Name")
	userMajor := c.Request.Header.Get("X-User-Major")

	if userId != "" && userRole != "" && userPermission != "" {

		//log.Println("Request Header is Valid")

		authenticated = &AuthenticatedRequest{
			Permissions: userPermission,
			UserID:      userId,
			Role:        userRole,
			Username:    username,
			UserMajor:   userMajor,
		}

		c.Set("authenticatedRequest", authenticated)
		c.Next()

	} else {
		log.Println("Request Header is Invalid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request Header is Invalid"})
		c.Abort()
	}

}
