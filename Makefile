SHELL := /bin/bash
VERSION := $(shell bump show meta/meta.go)

.PHONY: version
version:
	@echo "evans: $(VERSION)"


.PHONY: glide
glide:
ifeq ($(shell which glide 2>/dev/null),)
	curl https://glide.sh/get | sh
endif

.PHONY: deps
deps: glide
	glide install

.PHONY: build
build: deps
	go build

.PHONY: test
test: unit-test e2e-test

.PHONY: unit-test
unit-test:
	go test -race $(shell glide novendor | grep -v tests)

.PHONY: e2e-test
e2e-test:
	go test -race ./tests/...

.PHONY: coverage
coverage:
	go tool cover -html=coverage.out

.PHONY: brew-update
brew-update:
	bash .circleci/scripts/entrypoint.bash $(VERSION)
