APP := organizations

.PHONY: run build fmt vet lint test migrate-up migrate-down compose-up compose-down tidy

APP := organizations

run:
	go run ./cmd/organizations

build:
	mkdir -p bin
	GO111MODULE=on go build -o bin/$(APP) ./cmd/organizations

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	@if [ -z "$(shell command -v golangci-lint 2>/dev/null)" ]; then \
		GOBIN=$$(go env GOPATH)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	$$(go env GOPATH)/bin/golangci-lint run

test:
	go test ./...

migrate-up:
	@if [ -z "$(shell command -v migrate 2>/dev/null)" ]; then \
		GOBIN=$$(go env GOPATH)/bin go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	migrate -path db/migrations -database "$$DB_CONNECTION" up

migrate-down:
	migrate -path db/migrations -database "$$DB_CONNECTION" down 1

compose-up:
	docker compose up -d

compose-down:
	docker compose down -v

tidy:
	go mod tidy
