package handlers

import "github.com/gin-gonic/gin"

// JSONError writes a unified error envelope.
func JSONError(c *gin.Context, status int, code, message string, details any) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"details": details,
		},
	})
}
