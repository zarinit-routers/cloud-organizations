package node

import (
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getJwtKey() ([]byte, error) {
	key := os.Getenv("JWT_SECURITY_KEY")
	if key == "" {
		return nil, fmt.Errorf("environment variable JWT_SECURITY_KEY is not set")

	}

	return []byte(key), nil
}

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {

		var request Request
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if !matchPassphrase() {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		key, err := getJwtKey()
		if err != nil {
			log.Error("Failed get JWT key", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"id":      request.NodeId,
			"groupId": request.GroupId,
		})
		tokenString, err := token.SignedString(key)
		if err != nil {
			log.Error("Failed sign JWT token", "error", err)
		}

		c.JSON(http.StatusOK, Response{
			Token: tokenString,
		})

	}
}

type Request struct {
	NodeId     string `json:"id" binding:"required"`
	GroupId    string `json:"groupId" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
}
type Response struct {
	Token string `json:"token"`
}

func matchPassphrase() bool {
	log.Warn("Not implemented passphrase matching")
	return true
}
