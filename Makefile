.PHONY: build test vet lint bundle

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

bundle:
	cd channel && bun install && bun run build.mjs

build: bundle
	go build -ldflags "-X main.version=$(VERSION)" -o bin/monocle ./cmd/monocle

install: bundle
	go install ./cmd/monocle

test:
	go test ./internal/...

vet:
	go vet ./...

lint: vet
	go build ./...

