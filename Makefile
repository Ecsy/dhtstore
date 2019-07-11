SHELL := /bin/bash

# ---- Environment variables ----

#var GOBIN: ## Set up the output binary directory.
GOBIN = $(shell pwd)/bin

#var GO: ## go executable path.
GO ?= go

#var CGO_ENABLED: ## Enable the use of cgo.
CGO_ENABLED ?= 0

#var DATE: ## Date registered for compiled program.
DATE ?= $(shell date +%FT%T%z)

-include version.mk

#var VER: ## Program build version, baked in compile time.
VER ?= ?

export GOPATH GOBIN CGO_ENABLED

env: ## Show env variables.
	@echo "export GOPATH=\"$(GOPATH)\""
	@echo "export GOBIN=\"$(GOBIN)\""
	@echo "export CGO_ENABLED=\"$(CGO_ENABLED)\""

# ---- Check ----
go-exsits: ; @which $(GO) > /dev/null
check:: go-exsits

GO_LD_FLAGS = --ldflags '-X main.version=$(VER) -X main.buildDate=$(DATE)'

# ---- Compile ----
define COMPILE
	$($(1)) $($(1)FLAGS) $<
endef

# ---- Compilers ----
GOINSTALL=$(GO) install -v
COMPILE_GO = $(call COMPILE,GOINSTALL)

# ---- Compiler Flags ----
GOINSTALLFLAGS+=$(GO_LD_FLAGS)

# ---- Targets ----
TARGETS = \
	dhtstore

$(TARGETS):: check

dhtstore:: src/cmd/dhtstore/dhtstore.go ## Compile dhtstore.
	$(COMPILE_GO)

# ---- Common ----
.PHONY: check all build clean help env $(TARGETS)

.DEFAULT_GOAL := build
all: build

build: ## Build all targets.
	$(MAKE) $(TARGETS)

clean: ## Clean the source repository.
	$(RM) -r bin/*

help: ## Generates this help message
	@grep -hE '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":+[^:]*## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

