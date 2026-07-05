.PHONY: fmt vet lint test build all check tidy wire gowrap generate tools \
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

tools:
	$(GO) mod download
	$(GO) tool -n gowrap
	@echo Go tools ready: gowrap

wire:
	cd internal/app && $(GO) run github.com/google/wire/cmd/wire

gowrap:
	cd pkg/runner && $(GO) tool gowrap gen -p meguri/pkg/runner -i Crawler -t templates/slog_debug.gotmpl -o crawler_with_debug_log.go
	cd pkg/runner && $(GO) tool gowrap gen -p meguri/pkg/runner -i Scraper -t templates/slog_debug.gotmpl -o scraper_with_debug_log.go
	cd pkg/runner && $(GO) tool gowrap gen -p meguri/pkg/runner -i RobotsFetcher -t templates/slog_debug.gotmpl -o robots_fetcher_with_debug_log.go
	cd pkg/runner && $(GO) tool gowrap gen -p meguri/pkg/runner -i UIConfigurator -t templates/slog_debug.gotmpl -o uiconfigurator_with_debug_log.go
	cd pkg/runner && $(GO) tool gowrap gen -p meguri/pkg/runner -i FetchLimiterPreparer -t templates/slog_debug.gotmpl -o fetch_limit_preparer_with_debug_log.go
	cd pkg/runner && $(GO) tool gowrap gen -p meguri/pkg/runner -i ChromiumShutdown -t templates/slog_debug.gotmpl -o chromium_shutdown_with_debug_log.go

generate: wire gowrap

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
