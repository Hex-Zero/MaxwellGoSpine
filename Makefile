APP_NAME?=maxwell-api
GO_FILES=$(shell go list ./...)

.PHONY: run build test lint fmt vet tidy migrate-up migrate-down

run:
	go run ./cmd/server

build:
	go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)" -o bin/$(APP_NAME) ./cmd/server

test:
	go test -race -count=1 ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

migrate-up:
	migrate -path migrations -database $$DB_DSN up

migrate-down:
	migrate -path migrations -database $$DB_DSN down 1
