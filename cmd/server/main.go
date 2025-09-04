package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zarinit-routers/cloud-organizaions/handlers/node"
	"github.com/zarinit-routers/cloud-organizaions/middleware/auth"
	"github.com/zarinit-routers/cloud-organizaions/middleware/implementation"
)

func main() {

	server := gin.Default()

	api := server.Group("/api/organizations")

	api.GET("/", auth.AuthMiddleware(), auth.AdminOnly(), implementation.NotImplemented())
	api.GET("/:id", auth.AuthMiddleware(), implementation.NotImplemented())
	api.POST("/new", auth.AuthMiddleware(), auth.AdminOnly(), implementation.NotImplemented())
	api.POST("/update", auth.AuthMiddleware(), auth.AdminOnly(), implementation.NotImplemented())
	api.POST("/generate-passphrase", auth.AuthMiddleware(), auth.AdminOnly(), implementation.NotImplemented())
	api.POST("/delete", auth.AuthMiddleware(), auth.AdminOnly(), implementation.NotImplemented())
	api.POST("/add-users", auth.AuthMiddleware(), auth.AdminOnly(), implementation.NotImplemented())
	api.POST("/remove-users", auth.AuthMiddleware(), auth.AdminOnly(), implementation.NotImplemented())

	api.POST("/authorize-node", node.AuthorizeNode())

	server.Run(":8080")
}
