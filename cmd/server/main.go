package main

import (
	"github.com/charmbracelet/log"

	"github.com/zarinit-routers/cloud-organizaions/pkg/config"
	"github.com/zarinit-routers/cloud-organizaions/pkg/server"
	"github.com/zarinit-routers/cloud-organizaions/pkg/storage/database"
)

func main() {

	if err := config.Load(); err != nil {
		log.Fatal("Failed load configuration", "error", err)
	}

	if err := database.Migrate(); err != nil {
		log.Fatal("Failed migrate database", "error", err)
	}

	srv := server.New()
	if err := srv.Run(server.Address()); err != nil {
		log.Fatal("Failed start server", "error", err)
	}
}
