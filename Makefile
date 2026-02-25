.PHONY: build dev test test-integration test-web test-all lint migrate setup-dkim docker-build clean help

BINARY_NAME=mailit
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*' -not -path './web/*')

# Build
build:
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mailit

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mailit
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/mailit

# Development
dev:
	docker compose -f docker-compose.dev.yml up -d postgres redis
	@echo "Waiting for postgres..."
	@sleep 2
	go run ./cmd/mailit serve

dev-docker:
	docker compose -f docker-compose.dev.yml up --build

# Test
test:
	go test -race -cover ./...

test-integration:
	go test -race -tags integration ./internal/repository/postgres/...

test-web:
	cd web && npx vitest run

test-all: test test-web

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Database
migrate-up:
	go run ./cmd/mailit migrate --up

migrate-down:
	go run ./cmd/mailit migrate --down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

# DKIM
setup-dkim:
	./scripts/generate-dkim.sh

# Setup (first run)
setup:
	./scripts/setup.sh

# Docker
docker-build:
	docker build -f deploy/docker/Dockerfile -t mailit-api .
	docker build -f web/Dockerfile -t mailit-web ./web

docker-push:
	docker push mailit-api
	docker push mailit-web

# Docker Compose
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

# Clean
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

# Help
help:
	@echo "MailIt - Self-hosted email platform"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the Go binary"
	@echo "  make dev            Run locally (requires Docker for postgres/redis)"
	@echo "  make dev-docker     Run everything in Docker with hot reload"
	@echo "  make test           Run tests"
	@echo "  make lint           Run linter"
	@echo "  make migrate-up     Run database migrations"
	@echo "  make migrate-down   Roll back last migration"
	@echo "  make setup-dkim     Generate DKIM keys and print DNS records"
	@echo "  make setup          First-run setup (admin account + DKIM)"
	@echo "  make docker-build   Build Docker images"
	@echo "  make up             Start all services"
	@echo "  make down           Stop all services"
	@echo "  make logs           Follow logs"
	@echo "  make clean          Remove build artifacts"
