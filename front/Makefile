.PHONY: help setup dev run build clean check test lint format lint-fix \
	go-test go-build go-tidy frontend-install frontend-dev frontend-build frontend-preview frontend-test frontend-lint frontend-format frontend-lint-fix

GO ?= go
NPM ?= npm
WAILS ?= wails3
FRONTEND_DIR ?= frontend

all: check

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Wails app targets:"
	@echo "  setup     Install frontend dependencies"
	@echo "  dev       Run Wails dev mode"
	@echo "  run       Run built app"
	@echo "  build     Build app package"
	@echo "  clean     Remove generated frontend dist"
	@echo ""
	@echo "Quality targets:"
	@echo "  check     lint + test"
	@echo "  lint      Run frontend lint"
	@echo "  fmt       Run frontend format + lint auto-fix (biome check --write)"
	@echo "  generate    Generate frontend bindings"
	@echo "  test      Run Go + frontend tests"
	@echo ""
	@echo "Frontend-only shortcuts:"
	@echo "  frontend-dev frontend-build frontend-preview frontend-test"

setup: frontend-install

frontend-install:
	$(NPM) --prefix $(FRONTEND_DIR) install

dev:
	$(WAILS) dev -config ./build/config.yml -port $${WAILS_VITE_PORT:-9245}

run:
	$(WAILS) run

build:
	$(WAILS) build -ldflags="-s -w" 

generate:
	$(WAILS) generate bindings -ts

clean:
	$(NPM) --prefix $(FRONTEND_DIR) run --if-present clean

check: lint test

test: go-test frontend-test

lint: frontend-lint

fmt: frontend-lint-fix

format: frontend-format

lint-fix: frontend-lint-fix

go-test:
	$(GO) test ./...

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

frontend-format:
	$(NPM) --prefix $(FRONTEND_DIR) run format

frontend-lint-fix:
	$(NPM) --prefix $(FRONTEND_DIR) run lint:fix
