SHELL := /bin/bash

.PHONY: glide
glide:
ifeq ($(shell which glide 2>/dev/null),)
	curl https://glide.sh/get | sh
endif

.PHONY: deps
deps: glide
	@glide install

.PHONY: build
build: deps
	@go build 

.PHONY: test
test:
	@for p in $(shell go list $(shell glide novendor)); do go test -v -race -coverprofile=coverage.out $p; done

.PHONY: coverage
coverage: 
	@go tool cover -html=coverage.out
