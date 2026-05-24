.PHONY: fmt vet lint test build all check tidy wire

GO ?= go
GOLANGCI_LINT ?= golangci-lint
PKG := ./...

all: check

check: fmt vet lint test

fmt:
	$(GO) fmt $(PKG)

vet:
	$(GO) vet $(PKG)

lint:
	$(GOLANGCI_LINT) run $(PKG)

test:
	$(GO) test -race -count=1 $(PKG)

build:
	$(GO) build -o bin/scraperbot ./cmd/scraperbot

tidy:
	$(GO) mod tidy

wire:
	cd internal/app && $(GO) run github.com/google/wire/cmd/wire
