package node

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthorizeNode provides a compatibility endpoint for node authorization.
// Returns 200 OK with a simple authorization decision payload.
// In this minimal implementation, any request is considered authorized.
func AuthorizeNode() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"authorized": true,
		})
	}
}
