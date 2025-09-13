# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

Project: cloud-organizations (Go, Gin)

Quick start
- Requires Go >= 1.24.
- Entrypoint: cmd/organizations/main.go
- Default port: 8060

Run locally
- Config is read from environment variables with the prefix RAS_. Supported keys:
  - RAS_SERVER_PORT (default 8092)
  - RAS_SERVER_MODE: debug|release (default release)
  - RAS_LOG_LEVEL: debug|info|warn|error (default info)
  - RAS_CLIENT_ADDRESS: CORS AllowedOrigin (default *)
- The auth middleware expects a JWT secret. Export JWT_SECURITY_KEY with your secret so authenticated routes work.

```bash path=null start=null
# Install deps (first run)
go mod download

# Run (debug mode on a custom port)
export RAS_SERVER_MODE=debug
export RAS_SERVER_PORT=8060
export RAS_LOG_LEVEL=debug
export RAS_CLIENT_ADDRESS=http://localhost:3000
export JWT_SECURITY_KEY={{JWT_SECURITY_KEY}}

go run ./cmd/organizations
```

Build
```bash path=null start=null
# Build a local binary
mkdir -p bin
GO111MODULE=on go build -o bin/organizations ./cmd/organizations
```

Lint and format
- This repo has no Makefile yet. Use the following common commands:
```bash path=null start=null
# Format
go fmt ./...
# Vet
go vet ./...
# Optional: install and run golangci-lint
# (install once; add a pinned version as needed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
$(go env GOPATH)/bin/golangci-lint run
```

Tests
- There are currently no tests in the repo, but once they are added use:
```bash path=null start=null
# Run all tests
go test ./...
# Run tests in one package
go test ./pkg/services/organizations
# Run a single test by name (regex)
go test ./pkg/services/organizations -run ^TestService_Create$ -v
```

Docker
```bash path=null start=null
# Build
docker build -t cloud-organizations:local .
# Run
docker run --rm -p 8060:8060 \
  -e RAS_SERVER_PORT=8060 \
  -e RAS_SERVER_MODE=release \
  -e RAS_LOG_LEVEL=info \
  -e RAS_CLIENT_ADDRESS=http://localhost:3000 \
  -e JWT_SECURITY_KEY={{JWT_SECURITY_KEY}} \
  cloud-organizations:local
```

Operational endpoints
- /healthz, /readyz — probes
- /metrics — Prometheus metrics

Authentication (HTTP)
- The API group /api/v1/organizations is protected by a JWT middleware; write operations require admin role via auth.AdminOnly().
- For development, you can bypass user mapping with header X-User-ID (UUID). In production, prefer real JWT claims.

High-level architecture
- Server layer: pkg/server/server.go initializes Gin (mode from RAS_SERVER_MODE), CORS (RAS_CLIENT_ADDRESS), request id middleware, metrics, health endpoints, and wires routes.
- Handlers layer: pkg/handlers/organizations handles HTTP JSON, validation, and translates to the service layer. Errors use a unified JSON envelope from pkg/handlers/errors.go.
- Service layer: pkg/services/organizations provides business logic. Today it is an in-memory store using a map with soft-delete semantics and auditing fields (created/updated by/at). Bulk operations are supported.
- Models: pkg/models contains domain models and DTOs.
- Middleware: pkg/middleware/requestid injects X-Request-ID; external auth middleware provides JWT checks.
- Logger: pkg/logger/setup configures global log level.

Route map (excerpt)
```go path=/Users/zakhar/Projects/Zarinit/organisations/pkg/server/server.go start=62
// Probes and metrics omitted above
// API
api := r.Group("/api/v1/organizations")
h := orgHandlers.NewHandlers(svc)
{
    api.POST("/", auth.Middleware(auth.AdminOnly()), h.Create)
    api.GET("/:id", auth.Middleware(), h.Get)
    api.GET("/", auth.Middleware(), h.List)
    api.PUT("/:id", auth.Middleware(auth.AdminOnly()), h.Replace)
    api.PATCH("/:id", auth.Middleware(auth.AdminOnly()), h.Patch)
    api.DELETE("/:id", auth.Middleware(auth.AdminOnly()), h.Delete)
    api.POST("/:id/restore", auth.Middleware(auth.AdminOnly()), h.Restore)

    api.POST("/bulk", auth.Middleware(auth.AdminOnly()), h.BulkCreate)
    api.PATCH("/bulk", auth.Middleware(auth.AdminOnly()), h.BulkUpdate)
    api.DELETE("/bulk", auth.Middleware(auth.AdminOnly()), h.BulkDelete)
}

// Compatibility endpoint
r.POST("/api/v1/organizations/authorize-node", node.AuthorizeNode())
```

Notes for Warp
- Prefer running go run ./cmd/organizations during iteration; rebuild only when packaging.
- Use RAS_* env vars to control mode, port, and CORS. Set JWT_SECURITY_KEY for authenticated endpoints.
- The service is currently in-memory and stateless; data resets on restart.
- Env aliases supported: PORT -> server.port, GIN_MODE -> server.mode. Default port is 8060.
- Local DX: see Makefile (run/build/test/lint/migrate) and docker-compose.yml (postgres + rabbitmq).
