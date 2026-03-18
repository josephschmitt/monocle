.PHONY: build test vet lint

build:
	go build -o bin/monocle ./cmd/monocle

install:
	go install ./cmd/monocle

test:
	go test ./internal/...

vet:
	go vet ./...

lint: vet
	go build ./...

