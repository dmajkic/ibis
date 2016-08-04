package ibis

import "github.com/gin-gonic/gin"

// AppRouter interface that must be implemented by user application to set routes
type AppRouter interface {
	SetRoutes(router *gin.Engine)
}

// AppAuthorizer interface should be implemented by app to support user login
type AppAuthorizer interface {
	LoginUser(c *gin.Context, user map[string]interface{}) error
}
