.PHONY: build test lint fmt vet sqlc-generate migrate-up migrate-down migrate-version

DB_USER     ?= postgres
DB_PASSWORD ?= postgres
DB_HOST     ?= localhost
DB_PORT     ?= 5432
DB_NAME     ?= ai_trial_dev
DATABASE_URL ?= pgx5://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_DIR = ./db/migrations
MIGRATE = go run -tags 'pgx5' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

sqlc-generate:
	sqlc generate

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version
