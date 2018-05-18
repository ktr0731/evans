SHELL := /bin/bash
VERSION := $(shell bump show meta/meta.go)

.PHONY: version
version:
	@echo "evans: $(VERSION)"


.PHONY: dep
dep:
ifeq ($(shell which dep 2>/dev/null),)
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif

.PHONY: deps
deps: dep
	dep ensure

.PHONY: build
build: deps
	go build

.PHONY: test
test: unit-test e2e-test

.PHONY: unit-test
unit-test:
	go test -race $(shell go list ./... | grep -v tests)

.PHONY: e2e-test
e2e-test:
	go test -race ./tests/...

.PHONY: coverage
coverage:
	go tool cover -html=coverage.out

.PHONY: brew-update
brew-update:
	bash .circleci/scripts/entrypoint.bash $(VERSION)
