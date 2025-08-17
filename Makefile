APP_NAME?=maxwell-api
GO_FILES=$(shell go list ./...)
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
COMMIT_MSG?=chore: auto commit

.PHONY: run build test lint fmt vet tidy migrate-up migrate-down check auto-commit hooks-install

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

check: tidy lint test build

# Auto commit & push current changes after successful build/test.
# Skips commit if no staged/unstaged changes exist.
auto-commit: check
	@set -e; \
	git add -A; \
	if git diff --cached --quiet; then echo "No changes to commit"; else \
	  echo "Committing to $(GIT_BRANCH)"; \
	  git commit -m "$(COMMIT_MSG)"; \
	  git push origin $(GIT_BRANCH); \
	fi

migrate-up:
	migrate -path migrations -database $$DB_DSN up

migrate-down:
	migrate -path migrations -database $$DB_DSN down 1

# Install git hooks from .githooks directory
hooks-install:
	git config core.hooksPath .githooks
	@echo "Git hooks path set to .githooks (pre-commit enabled)"
