SHELL := /bin/bash

export GOBIN := $(PWD)/_tools
export PATH := $(GOBIN):$(PATH)
export GO111MODULE := on

.PHONY: version
version:
	@echo "evans: $(shell bump show meta/meta.go)"


.PHONY: dep
dep:
ifeq ($(shell go help mod 2>/dev/null),)
	@echo "Go v1.11 or later required"
	@exit 1
endif

.PHONY: deps
deps: dep
	@go mod download
	@go mod verify
	@go mod tidy

.PHONY: tools
tools:
	@cat tools/tools.go | grep -E '^\s*_\s.*' | awk '{ print $$2 }' | xargs go install

.PHONY: build
build: deps
	go build

.PHONY: build-dev
build-dev: deps
	go build -tags dev

.PHONY: test
test: format gotest

.PHONY: format
format:
	go mod tidy

.PHONY: gotest
gotest: lint
	go test -race ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: brew-update
release:
	bash scripts/release.bash $(shell bump show meta/meta.go)

.PHONY: depgraph
depgraph:
	GO111MODULE=off godepgraph -s -novendor . | dot -Tpng -o dep.png
