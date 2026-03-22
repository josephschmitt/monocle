.PHONY: build test vet lint

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/monocle ./cmd/monocle

install:
	go install ./cmd/monocle

test:
	go test ./internal/...

vet:
	go vet ./...

lint: vet
	go build ./...

