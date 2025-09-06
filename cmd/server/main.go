package main

import (
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/zarinit-routers/cloud-organizaions/handlers/node"
	"github.com/zarinit-routers/cloud-organizaions/middleware/implementation"
	"github.com/zarinit-routers/middleware/auth"
)

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

	log.Warn("Port 8060 is hardcoded, remove this ASAP")
	server.Run(":8060")
}
