SHELL := /bin/bash
VERSION := $(shell _tools/bump show meta/meta.go)

export PATH := $(PWD)/_tools:$(PATH)
export GO111MODULE := on

.PHONY: version
version:
	@echo "evans: $(VERSION)"

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
test: format unit-test e2e-test

.PHONY: format
format:
	go mod tidy

.PHONY: unit-test
unit-test: lint
	go test -race ./...

.PHONY: e2e-test
e2e-test: lint
	go test -tags e2e -race ./e2e/...

.PHONY: lint
lint:
	golangci-lint run --disable-all \
		--build-tags e2e \
		-E 'deadcode,govet,golint' \
		./...

.PHONY: coverage
coverage:
	go test -coverpkg ./... -covermode=atomic -coverprofile=coverage.txt -race ./...

.PHONY: coverage-circleci
coverage-circleci:
	go test -p 1 -coverpkg ./... -covermode=atomic -coverprofile=coverage.txt ./...

.PHONY: coverage-web
coverage-web: coverage
	go tool cover -html=coverage.txt

.PHONY: brew-update
release:
	bash .circleci/scripts/release.bash $(VERSION)

.PHONY: depgraph
depgraph:
	godepgraph -s -novendor . | dot -Tpng -o dep.png
