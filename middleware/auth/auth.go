package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	jwt "github.com/golang-jwt/jwt/v5"
)

const AUTH_DATA_KEY = "middleware-auth-data"

func getJwtKey() jwt.Keyfunc {

	return func(t *jwt.Token) (any, error) {
		key := os.Getenv("JWT_SECURITY_KEY")
		if key == "" {
			return nil, fmt.Errorf("environment variable JWT_SECURITY_KEY not specified")
		}
		return key, nil
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			log.Error("Authentication failed with no token specified in Authorization header")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(tokenString, getJwtKey())
		if err != nil {
			log.Error("Authentication failed", "error", err)
		}

		data := NewDataFromToken(token)

		c.Set(AUTH_DATA_KEY, data)

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		param, exists := c.Get(AUTH_DATA_KEY)
		if !exists {
			log.Error("Authentication failed with no data in current context")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		data := param.(*AuthData)
		if !data.IsAdmin() {
			log.Error("Authenticated user is not an admin")
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		c.Next()
	}
}
