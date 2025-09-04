package implementation

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Not implemented"})
	}
}
