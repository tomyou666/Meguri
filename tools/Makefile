.PHONY: help all check fmt lint vuln upgrade-patch upgrade-minor \
	$(addprefix fmt-,$(MODULES)) $(addprefix lint-,$(MODULES)) \
	$(addprefix gvuln-,$(MODULES)) \
	$(addprefix gupgrade-patch-,$(MODULES)) \
	$(addprefix gupgrade-minor-,$(MODULES))

GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOVULNCHECK ?= govulncheck

# go.mod を持つサブディレクトリを対象にする（runcomfy など Go 以外は除外）
MODULES := $(sort $(patsubst %/,%,$(dir $(wildcard */go.mod))))

all: check

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  check         Run fmt and lint on all tool modules"
	@echo "  fmt           Run go fmt on all tool modules"
	@echo "  lint          Run golangci-lint on all tool modules"
	@echo "  vuln          Run govulncheck on all tool modules"
	@echo "  upgrade-patch Upgrade patch versions for all tool modules"
	@echo "  upgrade-minor Upgrade minor versions for all tool modules"
	@echo ""
	@echo "Modules: $(MODULES)"

check: fmt lint

fmt: $(addprefix fmt-,$(MODULES))

fmt-%:
	@echo "==> $*: go fmt"
	$(GO) -C $* fmt ./...

lint: $(addprefix lint-,$(MODULES))

lint-%:
	@echo "==> $*: golangci-lint"
	cd $* && $(GOLANGCI_LINT) run ./...

gvuln: $(addprefix gvuln-,$(MODULES))

gvuln-%:
	@echo "==> $*: govulncheck"
	$(GOVULNCHECK) -C $* ./...

gupgrade-patch: $(addprefix gupgrade-patch-,$(MODULES))

gupgrade-patch-%:
	@echo "==> $*: go get -u=patch"
	$(GO) -C $* get -u=patch ./...
	$(GO) -C $* mod tidy

gupgrade-minor: $(addprefix gupgrade-minor-,$(MODULES))

gupgrade-minor-%:
	@echo "==> $*: go get -u"
	$(GO) -C $* get -u ./...
	$(GO) -C $* mod tidy

vuln: gvuln

upgrade-patch: gupgrade-patch

upgrade-minor: gupgrade-minor
