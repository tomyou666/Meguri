.PHONY: help all check fmt lint \
	$(addprefix fmt-,$(MODULES)) $(addprefix lint-,$(MODULES))

GO ?= go
GOLANGCI_LINT ?= golangci-lint

# go.mod を持つサブディレクトリを対象にする（runcomfy など Go 以外は除外）
MODULES := $(sort $(patsubst %/,%,$(dir $(wildcard */go.mod))))

all: check

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  check  Run fmt and lint on all tool modules"
	@echo "  fmt    Run go fmt on all tool modules"
	@echo "  lint   Run golangci-lint on all tool modules"
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
