.PHONY: help all check fmt vet lint test build tidy wire generate tools \
	vuln upgrade-patch upgrade-minor \
	bcheck bbuild btidy bwire bgenerate blint bfmt bvuln bupgrade-patch bupgrade-minor \
	fdev fbuild frun ftest flint ffmt fcheck fsetup ftools ftidy fwire fgenerate fvuln fupgrade-patch fupgrade-minor \
	tlint tfmt ttidy tvuln tupgrade-patch tupgrade-minor

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
	@echo "  tidy          Run go mod tidy on backend, front, and tools"
	@echo "  wire          Regenerate backend and front wire_gen.go"
	@echo "  generate      Regenerate all codegen (backend wire + front migrate/gen/wire/bindings)"
	@echo "  tools         Download front Go tool dependencies (dlv, migrate, wails3)"
	@echo "  vuln          Run govulncheck (all Go modules) + npm audit"
	@echo "  upgrade-patch Upgrade patch versions (Go + npm)"
	@echo "  upgrade-minor Upgrade minor versions (Go + npm)"
	@echo ""
	@echo "Backend shortcuts:"
	@echo "  bcheck bbuild btidy bwire bgenerate blint bfmt"
	@echo ""
	@echo "Front shortcuts:"
	@echo "  fsetup fdev fbuild frun ftidy fwire fgenerate"
	@echo "  ftest flint ffmt fcheck"
	@echo ""
	@echo "Tools shortcuts:"
	@echo "  tlint tfmt ttidy"

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

tidy: btidy ftidy ttidy

wire: bwire fwire

generate: bgenerate fgenerate

tools: ftools

vuln: bvuln fvuln tvuln

upgrade-patch: bupgrade-patch fupgrade-patch tupgrade-patch

upgrade-minor: bupgrade-minor fupgrade-minor tupgrade-minor

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

bgenerate:
	$(MAKE) -C $(BACKEND_DIR) generate

blint:
	$(MAKE) -C $(BACKEND_DIR) lint

bfmt:
	$(MAKE) -C $(BACKEND_DIR) fmt

bvuln:
	$(MAKE) -C $(BACKEND_DIR) gvuln

bupgrade-patch:
	$(MAKE) -C $(BACKEND_DIR) gupgrade-patch

bupgrade-minor:
	$(MAKE) -C $(BACKEND_DIR) gupgrade-minor

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

ftidy:
	$(MAKE) -C $(FRONT_DIR) tidy

fwire:
	$(MAKE) -C $(FRONT_DIR) wire

fgenerate:
	$(MAKE) -C $(FRONT_DIR) generate

fvuln:
	$(MAKE) -C $(FRONT_DIR) vuln

fupgrade-patch:
	$(MAKE) -C $(FRONT_DIR) upgrade-patch

fupgrade-minor:
	$(MAKE) -C $(FRONT_DIR) upgrade-minor

tlint:
	$(MAKE) -C $(TOOLS_DIR) lint

tfmt:
	$(MAKE) -C $(TOOLS_DIR) fmt

ttidy:
	$(MAKE) -C $(TOOLS_DIR) tidy

tvuln:
	$(MAKE) -C $(TOOLS_DIR) gvuln

tupgrade-patch:
	$(MAKE) -C $(TOOLS_DIR) gupgrade-patch

tupgrade-minor:
	$(MAKE) -C $(TOOLS_DIR) gupgrade-minor
