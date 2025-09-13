package main

import (
	"os"

	"github.com/charmbracelet/log"

	"github.com/zarinit-routers/cloud-organizations/pkg/config"
	"github.com/zarinit-routers/cloud-organizations/pkg/server"
	orgsvc "github.com/zarinit-routers/cloud-organizations/pkg/services/organizations"
)

func main() {
	if err := config.Load(); err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	log.Info("starting service")

	// In-memory service for now (Sprint 1 keeps service simple)
	svc := orgsvc.NewService(nil) // publisher can be wired later
	r := server.Server(svc)

	addr := server.Addr()
	log.Info("listening", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Error("server exited", "error", err)
		os.Exit(1)
	}
}
