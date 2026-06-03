.PHONY: help all check fmt vet lint test build tidy wire \
	backend-check backend-build backend-tidy backend-wire \
	front-dev front-build front-run front-test front-lint front-format front-check front-setup

BACKEND_DIR ?= backend
FRONT_DIR ?= front

all: check

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Project targets:"
	@echo "  check         Run backend and front checks"
	@echo "  build         Build backend and front app"
	@echo "  test          Run backend and front tests"
	@echo "  lint          Run backend and front lint"
	@echo "  fmt           Run backend fmt and front lint:fix"
	@echo "  tidy          Run backend go mod tidy"
	@echo "  wire          Regenerate backend wire_gen.go"
	@echo ""
	@echo "Backend shortcuts:"
	@echo "  backend-check backend-build backend-tidy backend-wire"
	@echo ""
	@echo "Front shortcuts:"
	@echo "  front-setup front-dev front-build front-run"
	@echo "  front-test front-lint front-format front-check"

check: backend-check front-check

build: backend-build front-build

test:
	$(MAKE) -C $(BACKEND_DIR) test
	$(MAKE) -C $(FRONT_DIR) test

lint:
	$(MAKE) -C $(BACKEND_DIR) lint
	$(MAKE) -C $(FRONT_DIR) lint

fmt:
	$(MAKE) -C $(BACKEND_DIR) fmt
	$(MAKE) -C $(FRONT_DIR) fmt

tidy: backend-tidy

wire: backend-wire

backend-check:
	$(MAKE) -C $(BACKEND_DIR) check

backend-build:
	$(MAKE) -C $(BACKEND_DIR) build

backend-tidy:
	$(MAKE) -C $(BACKEND_DIR) tidy

backend-wire:
	$(MAKE) -C $(BACKEND_DIR) wire

front-setup:
	$(MAKE) -C $(FRONT_DIR) setup

front-dev:
	$(MAKE) -C $(FRONT_DIR) dev

front-build:
	$(MAKE) -C $(FRONT_DIR) build

front-run:
	$(MAKE) -C $(FRONT_DIR) run

front-test:
	$(MAKE) -C $(FRONT_DIR) test

front-lint:
	$(MAKE) -C $(FRONT_DIR) lint

front-format:
	$(MAKE) -C $(FRONT_DIR) format

front-check:
	$(MAKE) -C $(FRONT_DIR) check
