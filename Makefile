SHELL := /bin/bash

.PHONY: glide
glide:
	@curl https://glide.sh/get | sh

.PHONY: deps
deps: glide
	@glide install

.PHONY: build
build: deps
	@go build 

.PHONY: test
test:
	@for p in $$(go list $$(glide novendor)); do go test -v -race -coverprofile=coverage.out $$p; done

.PHONY: coverage
coverage: 
	@go tool cover -html=coverage.out
