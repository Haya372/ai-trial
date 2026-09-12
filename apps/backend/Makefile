.PHONY: build test lint fmt vet sqlc-generate oapi-codegen-generate migrate-up migrate-down migrate-version

DB_USER     ?= postgres
DB_PASSWORD ?= postgres
DB_HOST     ?= localhost
DB_PORT     ?= 5432
DB_NAME     ?= ai_trial_dev
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_DIR = ./db/migrations
MIGRATE = mise exec -- migrate

build:
	go build ./...

test:
	go test $(shell go list ./... | grep -v generated)

lint:
	mise exec -- golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

sqlc-generate:
	sqlc generate

oapi-codegen-generate:
	go tool oapi-codegen -config ../../packages/schema/oapi-codegen.yaml ../../packages/schema/dist/openapi.yaml

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version
