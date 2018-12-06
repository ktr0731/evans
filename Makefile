SHELL := /bin/bash
VERSION := $(shell bump show meta/meta.go)

export PATH := _tools:$(PATH)
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
	@go mod vendor
	@go get github.com/ktr0731/dept@v0.1.0
	@go build -o _tools/dept github.com/ktr0731/dept
	@dept build

.PHONY: generate
generate:
ifneq ($(shell git diff entity),)
	moq -pkg mockentity -out tests/mock/entity/mock.go entity Field Message Service RPC ClientStream ServerStream BidiStream GRPCClient
endif

ifneq ($(shell git diff entity/env),)
	moq -pkg mockenv -out tests/mock/entity/mockenv/mock.go entity/env Environment
endif

ifneq ($(shell git diff usecase/port),)
	moq -pkg mockport -out tests/mock/usecase/mockport/mock.go usecase/port InputPort Showable Inputter OutputPort DynamicBuilder
endif

.PHONY: build
build: deps
	go build

.PHONY: test
test: format unit-test e2e-test

.PHONY: format
format:
	go mod tidy

.PHONY: unit-test
unit-test: lint
	go test -v -race ./...

.PHONY: e2e-test
e2e-test: lint
	go test -v -tags e2e -race ./tests/...

.PHONY: lint
lint:
	golangci-lint run --disable-all \
		--build-tags e2e \
		-e 'should have name of the form ErrFoo' -E 'deadcode,govet,golint' \
		./...

.PHONY: coverage
coverage:
	go test -coverpkg ./... -covermode=atomic -tags e2e -coverprofile=coverage.txt -race $(shell go list ./... | egrep -v '(mock|testentity|helper|di)')

.PHONY: coverage-web
coverage-web: coverage
	go tool cover -html=coverage.txt

.PHONY: brew-update
brew-update:
	bash .circleci/scripts/entrypoint.bash $(VERSION)

.PHONY: depgraph
depgraph:
	godepgraph -s -novendor . | dot -Tpng -o dep.png
