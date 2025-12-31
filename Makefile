BINARY_NAME=gsc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X gsc-cli/internal/cmd.Version=$(VERSION) \
                  -X gsc-cli/internal/cmd.Commit=$(COMMIT) \
                  -X gsc-cli/internal/cmd.Date=$(DATE)"

.PHONY: all build install clean test lint dev

all: build

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/gsc

install:
	go install $(LDFLAGS) ./cmd/gsc

clean:
	rm -rf bin/
	go clean

test:
	go test -v ./...

lint:
	golangci-lint run

dev:
	go run ./cmd/gsc

# Run with arguments (e.g., make run ARGS="queries --days 7")
run:
	go run ./cmd/gsc $(ARGS)

# Cross-compilation targets
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/gsc
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/gsc
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/gsc
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/gsc

# Generate shell completions
completions:
	mkdir -p completions
	bin/$(BINARY_NAME) completion bash > completions/gsc.bash
	bin/$(BINARY_NAME) completion zsh > completions/_gsc
	bin/$(BINARY_NAME) completion fish > completions/gsc.fish
