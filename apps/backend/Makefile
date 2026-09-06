.PHONY: build test lint fmt vet

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
