APP_NAME := docuguard
VERSION := $(shell git describe --tags --always --dirty 2>nul || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>nul || echo "unknown")
BUILD_DATE := $(shell powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'" 2>nul || date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "-X github.com/blueberrycongee/docuguard/cli.Version=$(VERSION) \
                     -X github.com/blueberrycongee/docuguard/cli.GitCommit=$(GIT_COMMIT) \
                     -X github.com/blueberrycongee/docuguard/cli.BuildDate=$(BUILD_DATE)"

.PHONY: build test lint clean install fmt

build:
	@if not exist bin mkdir bin
	go build $(LDFLAGS) -o bin/$(APP_NAME).exe ./cmd/docuguard

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	@if exist bin rmdir /s /q bin
	@if exist dist rmdir /s /q dist
	@if exist coverage.out del /q coverage.out

install:
	go install $(LDFLAGS) ./cmd/docuguard

fmt:
	go fmt ./...
