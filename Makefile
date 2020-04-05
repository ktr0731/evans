SHELL := /bin/bash

export PATH := $(PWD)/_tools:$(PATH)
export GO111MODULE := on

.PHONY: version
version:
	@echo "evans: $(shell _tools/bump show meta/meta.go)"


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

.PHONY: dept
dept:
	@go get github.com/ktr0731/dept@v0.1.1
	@go build -o _tools/dept github.com/ktr0731/dept

.PHONY: tools
tools: dept
	@dept -v build

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
	bash .circleci/scripts/release.bash $(shell _tools/bump show meta/meta.go)

.PHONY: depgraph
depgraph:
	GO111MODULE=off godepgraph -s -novendor . | dot -Tpng -o dep.png
