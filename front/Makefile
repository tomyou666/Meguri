.PHONY: help setup dev run build clean check test lint format lint-fix \
	go-test go-build go-tidy go-fmt go-vet go-lint \
	frontend-install frontend-dev frontend-build frontend-preview frontend-test frontend-lint frontend-typecheck frontend-format frontend-lint-fix \
	tools migrate-up migrate-down migrate-version gen wire bindings generate

GO ?= go
GOLANGCI_LINT ?= golangci-lint
NPM ?= npm
WAILS ?= $(GO) tool wails3
FRONTEND_DIR ?= frontend

DB ?= data/scraperbot.db
MIGRATIONS_PATH ?= internal/app/migrations
MIGRATE_DATABASE ?= sqlite://$(DB)
MIGRATE ?= $(GO) tool migrate

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
	@echo "  generate        Alias for bindings"
	@echo ""
	@echo "Quality targets:"
	@echo "  check     lint + test"
	@echo "  lint      Run golangci-lint + frontend lint + typecheck"
	@echo "  fmt       Run go fmt + frontend lint auto-fix"
	@echo "  test      Run Go + frontend tests"
	@echo "  go-fmt    Run go fmt on Go packages"
	@echo "  go-lint   Run golangci-lint on Go packages"
	@echo ""
	@echo "Frontend-only shortcuts:"
	@echo "  frontend-dev frontend-build frontend-preview frontend-test"

setup: frontend-install

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

generate: bindings

frontend-install:
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

test: go-test frontend-test

lint: go-lint frontend-lint frontend-typecheck

fmt: go-fmt frontend-lint-fix

go-fmt:
	$(GO) fmt ./...

go-vet:
	$(GO) vet -tags dev ./...

go-lint:
	$(GOLANGCI_LINT) run ./...

format: frontend-format

lint-fix: frontend-lint-fix

go-test:
	$(GO) test $(GO_TEST_FLAGS) -tags dev ./...

go-build:
	$(GO) build ./...

go-tidy:
	$(GO) mod tidy

frontend-dev:
	$(NPM) --prefix $(FRONTEND_DIR) run dev

frontend-build:
	$(NPM) --prefix $(FRONTEND_DIR) run build

frontend-preview:
	$(NPM) --prefix $(FRONTEND_DIR) run preview

frontend-test:
	$(NPM) --prefix $(FRONTEND_DIR) run test

frontend-lint:
	$(NPM) --prefix $(FRONTEND_DIR) run lint

frontend-typecheck:
	$(NPM) --prefix $(FRONTEND_DIR) run typecheck

frontend-format:
	$(NPM) --prefix $(FRONTEND_DIR) run format

frontend-lint-fix:
	$(NPM) --prefix $(FRONTEND_DIR) run lint:fix
