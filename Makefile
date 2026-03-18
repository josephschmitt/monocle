.PHONY: build test vet lint

build:
	go build -o bin/monocle ./cmd/monocle

test:
	go test ./internal/...

vet:
	go vet ./...

lint: vet
	go build ./...

