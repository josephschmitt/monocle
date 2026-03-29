.PHONY: build run test vet lint bundle sync-skills

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

bundle:
	cd channel && bun install && bun run build.mjs

build: bundle
	go build -ldflags "-X main.version=$(VERSION)" -o bin/monocle ./cmd/monocle

run: build
	./bin/monocle

install: bundle
	go install ./cmd/monocle

test:
	go test ./internal/...

vet:
	go vet ./...

lint: vet
	go build ./...

SKILL_NAMES := $(notdir $(patsubst %/SKILL.md,%,$(wildcard skills/*/SKILL.md)))
PLUGIN_AGENTS := claude codex gemini

sync-skills:
	@for agent in $(PLUGIN_AGENTS); do \
		rm -rf plugins/$$agent/skills; \
		mkdir -p plugins/$$agent/skills; \
		for skill in $(SKILL_NAMES); do \
			cp -r skills/$$skill plugins/$$agent/skills/$$skill; \
		done; \
	done

