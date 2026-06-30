.PHONY: fmt vet lint test build all check tidy wire \
	gvuln gupgrade-patch gupgrade-minor vuln upgrade-patch upgrade-minor

GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOVULNCHECK ?= govulncheck
PKG := ./...

# -race requires CGO; Windows disables CGO by default.
ifeq ($(OS),Windows_NT)
GO_TEST_FLAGS ?= -count=1
else
GO_TEST_FLAGS ?= -race -count=1
endif

all: check

check: fmt vet lint test

fmt:
	$(GO) fmt $(PKG)

vet:
	$(GO) vet $(PKG)

lint:
	$(GOLANGCI_LINT) run $(PKG)

test:
	$(GO) test $(GO_TEST_FLAGS) $(PKG)

build:
	$(GO) build -o bin/meguri-cli ./cmd/meguri-cli

tidy:
	$(GO) mod tidy

wire:
	cd internal/app && $(GO) run github.com/google/wire/cmd/wire

gvuln:
	$(GOVULNCHECK) ./...

gupgrade-patch:
	$(GO) get -u=patch ./...
	$(GO) mod tidy

gupgrade-minor:
	$(GO) get -u ./...
	$(GO) mod tidy

vuln: gvuln

upgrade-patch: gupgrade-patch

upgrade-minor: gupgrade-minor
