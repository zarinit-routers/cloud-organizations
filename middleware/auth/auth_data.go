package auth

import (
	"slices"

	"github.com/golang-jwt/jwt/v5"
)

type AuthData struct {
	UserId  string
	GroupId string
	Groups  []string
}

const AdminGroup = "admin"

func (a *AuthData) IsAdmin() bool {
	return slices.Contains(a.Groups, AdminGroup)
}

func NewDataFromToken(t *jwt.Token) *AuthData {
	return &AuthData{
		UserId:  t.Claims.(jwt.MapClaims)["userId"].(string),
		GroupId: t.Claims.(jwt.MapClaims)["groupId"].(string),
		Groups:  t.Claims.(jwt.MapClaims)["groups"].([]string),
	}
}
