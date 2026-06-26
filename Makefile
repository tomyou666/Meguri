.PHONY: help all check fmt vet lint test build tidy wire tools \
	bcheck bbuild btidy bwire \
	fdev fbuild frun ftest flint ffmt fcheck fsetup ftools \
	tlint tfmt

BACKEND_DIR ?= backend
FRONT_DIR ?= front
TOOLS_DIR ?= tools

all: check

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Project targets:"
	@echo "  check         Run backend and front checks"
	@echo "  build         Build backend and front app"
	@echo "  test          Run backend and front tests"
	@echo "  lint          Run backend, front, and tools lint"
	@echo "  fmt           Run backend, front, and tools fmt"
	@echo "  tidy          Run backend go mod tidy"
	@echo "  wire          Regenerate backend wire_gen.go"
	@echo "  tools         Download front Go tool dependencies (dlv, migrate, wails3)"
	@echo ""
	@echo "Backend shortcuts:"
	@echo "  bcheck bbuild btidy bwire"
	@echo ""
	@echo "Front shortcuts:"
	@echo "  fsetup fdev fbuild frun"
	@echo "  ftest flint ffmt fcheck"
	@echo ""
	@echo "Tools shortcuts:"
	@echo "  tlint tfmt"

check: bcheck fcheck

build: bbuild fbuild

test:
	$(MAKE) -C $(BACKEND_DIR) test
	$(MAKE) -C $(FRONT_DIR) test

lint:
	$(MAKE) -C $(BACKEND_DIR) lint
	$(MAKE) -C $(FRONT_DIR) lint
	$(MAKE) -C $(TOOLS_DIR) lint

fmt:
	$(MAKE) -C $(BACKEND_DIR) fmt
	$(MAKE) -C $(FRONT_DIR) fmt
	$(MAKE) -C $(TOOLS_DIR) fmt

tidy: btidy

wire: bwire

tools: ftools

ftools:
	$(MAKE) -C $(FRONT_DIR) tools

bcheck:
	$(MAKE) -C $(BACKEND_DIR) check

bbuild:
	$(MAKE) -C $(BACKEND_DIR) build

btidy:
	$(MAKE) -C $(BACKEND_DIR) tidy

bwire:
	$(MAKE) -C $(BACKEND_DIR) wire

fsetup:
	$(MAKE) -C $(FRONT_DIR) setup

fdev:
	$(MAKE) -C $(FRONT_DIR) dev

fbuild:
	$(MAKE) -C $(FRONT_DIR) build

frun:
	$(MAKE) -C $(FRONT_DIR) run

ftest:
	$(MAKE) -C $(FRONT_DIR) test

flint:
	$(MAKE) -C $(FRONT_DIR) lint

ffmt:
	$(MAKE) -C $(FRONT_DIR) fmt

fcheck:
	$(MAKE) -C $(FRONT_DIR) check

tlint:
	$(MAKE) -C $(TOOLS_DIR) lint

tfmt:
	$(MAKE) -C $(TOOLS_DIR) fmt
