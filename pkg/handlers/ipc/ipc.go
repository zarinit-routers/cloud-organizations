package ipc

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/zarinit-routers/cloud-organizaions/pkg/models"
	"github.com/zarinit-routers/cloud-organizaions/pkg/storage/database"
)

type Request struct {
	UserID uuid.UUID `json:"id"`
}

func GetOrganizationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("Authorization header is missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		_, err := jwt.Parse(authHeader, func(token *jwt.Token) (any, error) {
			return []byte(viper.GetString("jwt-security-key")), nil
		})
		if err != nil {
			log.Error("Failed to parse token", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var request Request
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Failed to bind request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		var member models.Member
		db := database.MustConnect()
		if err := db.Model(&models.Member{}).First(&member, request.UserID); err != nil {
			log.Error("Failed to get member", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"organizationId": member.OrganizationID})
	}

}
