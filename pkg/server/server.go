package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/zarinit-routers/cloud-organizaions/pkg/handlers/node"
	"github.com/zarinit-routers/cloud-organizaions/pkg/handlers/organizations"
	"github.com/zarinit-routers/middleware/auth"
)

func New() *gin.Engine {
	server := gin.Default()

	api := server.Group("/api/organizations")

	api.GET("/", auth.Middleware(auth.AdminOnly()), organizations.ListHandler())
	api.GET("/:id", auth.Middleware(), organizations.GetHandler())
	api.POST("/new", auth.Middleware(auth.AdminOnly()), organizations.NewHandler())
	api.POST("/update", auth.Middleware(auth.AdminOnly()), organizations.UpdateHandler())
	api.POST("/generate-passphrase", auth.Middleware(auth.AdminOnly()), organizations.GeneratePassphraseHandler())
	api.POST("/delete", auth.Middleware(auth.AdminOnly()), organizations.DeleteHandler())
	api.POST("/add-users", auth.Middleware(auth.AdminOnly()), organizations.AddMembersHandler())
	api.POST("/remove-users", auth.Middleware(auth.AdminOnly()), organizations.RemoveMembersHandler())

	api.POST("/authorize-node", node.Authorize())

	return server
}

const DefaultPort = 8060

func getPort() int {
	viper.SetDefault("port", DefaultPort)
	port := viper.GetInt("port")
	return port
}

func Address() string {
	return fmt.Sprintf(":%d", getPort())
}
