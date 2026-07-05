.PHONY: help setup dev run build clean check test lint fmt tidy lintfix \
	gtest gbuild gtidy gfmt gvet glint gvuln gupgrade-patch gupgrade-minor \
	finstall fdev fbuild fpreview ftest flint ftypecheck ffmt flintfix \
	faudit fupgrade-patch fupgrade-minor vuln upgrade-patch upgrade-minor \
	tools migrate-up migrate-down migrate-version gen wire bindings generate

GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOVULNCHECK ?= govulncheck
NPM ?= npm
NPM_AUDIT_LEVEL ?= high
NCU ?= npx --yes npm-check-updates
WAILS ?= $(GO) tool wails3
FRONTEND_DIR ?= frontend

DB ?= data/scraperbot.db
MIGRATIONS_PATH ?= internal/app/migrations
MIGRATE_DATABASE ?= sqlite://$(DB)
MIGRATE ?= $(GO) tool migrate

# main (//go:embed frontend/dist) はビルド済みアセットが無いと go list ./... が失敗するため除外する。
GTEST_DIRS := ./internal/... ./cmd/... ./shared/... ./tools/... ./build/...

all: check

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Wails app targets:"
	@echo "  setup       Install frontend dependencies"
	@echo "  dev         Run Wails dev mode"
	@echo "  run         Run built app"
	@echo "  build       Build app package"
	@echo "  clean       Remove generated frontend dist"
	@echo ""
	@echo "Codegen / DB targets:"
	@echo "  tools           Download Go tool dependencies (dlv, migrate, wails3)"
	@echo "  migrate-up      Apply pending SQL migrations"
	@echo "  migrate-down    Roll back one migration"
	@echo "  migrate-version Show current DB migration version"
	@echo "  gen             GORM Gen (requires migrate-up first)"
	@echo "  wire            Regenerate internal/app/wire_gen.go"
	@echo "  bindings        Regenerate frontend/bindings (TypeScript)"
	@echo "  generate        migrate-up + gen + wire + bindings"
	@echo ""
	@echo "Quality targets:"
	@echo "  check     lint + test"
	@echo "  lint      Run golangci-lint + frontend lint + typecheck"
	@echo "  fmt       Run go fmt + frontend lint auto-fix"
	@echo "  tidy      Run go mod tidy"
	@echo "  test      Run Go + frontend tests"
	@echo "  gfmt      Run go fmt on Go packages"
	@echo "  glint     Run golangci-lint on Go packages"
	@echo ""
	@echo "Dependency targets:"
	@echo "  vuln            Run govulncheck + npm audit"
	@echo "  gvuln           Run govulncheck on Go packages"
	@echo "  faudit          Run npm audit"
	@echo "  upgrade-patch   Upgrade Go (patch) + npm (patch) dependencies"
	@echo "  upgrade-minor   Upgrade Go (minor) + npm (minor) dependencies"
	@echo "  gupgrade-patch  Upgrade Go dependencies (patch only)"
	@echo "  gupgrade-minor  Upgrade Go dependencies (minor + patch)"
	@echo "  fupgrade-patch  Upgrade npm dependencies (patch only)"
	@echo "  fupgrade-minor  Upgrade npm dependencies (minor + patch)"
	@echo ""
	@echo "Frontend-only shortcuts:"
	@echo "  fdev fbuild fpreview ftest"

setup: finstall

tools:
	$(GO) mod download
	$(GO) tool -n dlv
	$(GO) tool -n migrate
	$(GO) tool -n wails3
	@echo Go tools ready: dlv, migrate, wails3

ifeq ($(OS),Windows_NT)
MIGRATE_MKDIR = powershell -NoProfile -Command "New-Item -ItemType Directory -Path (Split-Path '$(DB)') -Force | Out-Null"
else
MIGRATE_MKDIR = mkdir -p $(dir $(DB))
endif

migrate-up:
	$(MIGRATE_MKDIR)
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(MIGRATE_DATABASE)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(MIGRATE_DATABASE)" down 1

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(MIGRATE_DATABASE)" version

gen:
	$(GO) run ./tools/gen

wire:
	cd internal/app && $(GO) run github.com/google/wire/cmd/wire

bindings:
	$(WAILS) generate bindings -ts

generate: migrate-up gen wire bindings

finstall:
	$(NPM) --prefix $(FRONTEND_DIR) install

dev:
	$(WAILS) dev -config ./build/config.yml -port $${WAILS_VITE_PORT:-9245}

run:
	$(WAILS) run

build:
	$(WAILS) build -ldflags="-s -w"

clean:
	$(NPM) --prefix $(FRONTEND_DIR) run --if-present clean

check: lint test

test: gtest ftest

lint: glint flint ftypecheck

fmt: gfmt flintfix

gfmt:
	$(GO) fmt ./...

gvet:
	$(GO) vet -tags dev ./...

glint:
	$(GOLANGCI_LINT) run ./...

lintfix: flintfix

gtest:
	$(GO) test $(GO_TEST_FLAGS) -tags dev $(GTEST_DIRS)

gbuild:
	$(GO) build ./...

tidy: gtidy

gtidy:
	$(GO) mod tidy

gvuln:
	$(GOVULNCHECK) ./...

gupgrade-patch:
	$(GO) get -u=patch ./...
	$(GO) mod tidy

gupgrade-minor:
	$(GO) get -u ./...
	$(GO) mod tidy

faudit:
	$(NPM) --prefix $(FRONTEND_DIR) audit --audit-level=$(NPM_AUDIT_LEVEL)

fupgrade-patch:
	cd $(FRONTEND_DIR) && $(NCU) -u --target patch
	$(NPM) --prefix $(FRONTEND_DIR) install
	-$(NPM) --prefix $(FRONTEND_DIR) audit fix

fupgrade-minor:
	cd $(FRONTEND_DIR) && $(NCU) -u --target minor
	$(NPM) --prefix $(FRONTEND_DIR) install

vuln: gvuln faudit

upgrade-patch: gupgrade-patch fupgrade-patch

upgrade-minor: gupgrade-minor fupgrade-minor

fdev:
	$(NPM) --prefix $(FRONTEND_DIR) run dev

fbuild:
	$(NPM) --prefix $(FRONTEND_DIR) run build

fpreview:
	$(NPM) --prefix $(FRONTEND_DIR) run preview

ftest:
	$(NPM) --prefix $(FRONTEND_DIR) run test

flint:
	$(NPM) --prefix $(FRONTEND_DIR) run lint

ftypecheck:
	$(NPM) --prefix $(FRONTEND_DIR) run typecheck

ffmt:
	$(NPM) --prefix $(FRONTEND_DIR) run format

flintfix:
	$(NPM) --prefix $(FRONTEND_DIR) run lint:fix
