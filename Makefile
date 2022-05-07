SHELL := /bin/bash

export GOBIN := $(PWD)/_tools
export PATH := $(GOBIN):$(PATH)

.PHONY: version
version:
	@echo "evans: $(shell bump show meta/meta.go)"

.PHONY: tools
tools:
	@cat tools/tools.go | grep -E '^\s*_\s.*' | awk '{ print $$2 }' | xargs go install

.PHONY: build
build:
	go build

.PHONY: build-dev
build-dev:
	go build -tags dev

.PHONY: test
test: format gotest

.PHONY: format
format:
	go mod tidy

.PHONY: credits
credits:
	gocredits -skip-missing . > CREDITS

.PHONY: gotest
gotest: lint
	# TODO: Remove conflictPolicy flag.
	go test -race -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: release
release:
	bash scripts/release.bash $(shell bump show meta/meta.go)

.PHONY: depgraph
depgraph:
	GO111MODULE=off godepgraph -s -novendor . | dot -Tpng -o dep.png
