.PHONY: build install clean test lint run help

# Variables
BINARY_NAME=gitscrum
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

## build: Build the CLI binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/gitscrum

## install: Install the CLI to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/gitscrum

## run: Run the CLI
run:
	go run $(LDFLAGS) ./cmd/gitscrum $(ARGS)

## test: Run tests
test:
	go test -v -race -cover ./...

## lint: Run linter
lint:
	golangci-lint run ./...

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -rf dist/

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## completion: Generate shell completions
completion:
	@mkdir -p completions
	go run ./cmd/gitscrum completion bash > completions/gitscrum.bash
	go run ./cmd/gitscrum completion zsh > completions/_gitscrum
	go run ./cmd/gitscrum completion fish > completions/gitscrum.fish
	go run ./cmd/gitscrum completion powershell > completions/gitscrum.ps1

## release: Create a new release (requires goreleaser)
release:
	goreleaser release --clean

## snapshot: Create a snapshot release (for testing)
snapshot:
	goreleaser release --snapshot --clean

## help: Show this help message
help:
	@echo "GitScrum CLI - Development Commands"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^##' Makefile | sed 's/## /  /'
