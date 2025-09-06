package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/zarinit-routers/cloud-organizaions/handlers/node"
	"github.com/zarinit-routers/cloud-organizaions/middleware/implementation"
	"github.com/zarinit-routers/middleware/auth"
)

func init() {
	viper.AutomaticEnv()
}
func main() {

	server := gin.Default()

	api := server.Group("/api/organizations")

	api.GET("/", auth.Middleware(auth.AdminOnly()), implementation.NotImplemented())
	api.GET("/:id", auth.Middleware(), implementation.NotImplemented())
	api.POST("/new", auth.Middleware(auth.AdminOnly()), implementation.NotImplemented())
	api.POST("/update", auth.Middleware(auth.AdminOnly()), implementation.NotImplemented())
	api.POST("/generate-passphrase", auth.Middleware(auth.AdminOnly()), implementation.NotImplemented())
	api.POST("/delete", auth.Middleware(auth.AdminOnly()), implementation.NotImplemented())
	api.POST("/add-users", auth.Middleware(auth.AdminOnly()), implementation.NotImplemented())
	api.POST("/remove-users", auth.Middleware(auth.AdminOnly()), implementation.NotImplemented())

	api.POST("/authorize-node", node.AuthorizeNode())

	server.Run(getAddress())
}

const DefaultPort = 8060

func getPort() int {
	viper.SetDefault("port", DefaultPort)
	port := viper.GetInt("port")
	return port
}

func getAddress() string {
	return fmt.Sprintf(":%d", getPort())
}
