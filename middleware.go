package ibis

import (
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
)

// SetMiddleware sets some basic API middleware
func (s *Server) SetMiddleware(router *gin.Engine) {
	router.Use(TracerMiddleware(s))
	router.Use(CORSMiddleware())
}

// TracerMiddleware adds session reference, uuid identification to every request
func TracerMiddleware(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := uuid.NewV4()

		c.Set("uuid", id)
		c.Set("Server", s)
	}
}

// CORSMiddleware allows cross-site requests using standard HTTP Access-Controll-* headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
		}
	}
}
