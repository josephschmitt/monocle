.PHONY: build test vet

build:
	go build -o bin/monocle ./cmd/monocle
	go build -o bin/monocle-hook ./cmd/monocle-hook

test:
	go test ./internal/...

vet:
	go vet ./...
