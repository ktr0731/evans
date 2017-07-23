SHELL := /bin/bash

.PHONY: deps
build:
	glide install

.PHONY: build
build:
	go build 

.PHONY: test
test:
	for p in `go list`; do go test -v -race -coverprofile=coverage.out $p; done

.PHONY: coverage
coverage: 
	go tool cover -html=coverage.out
