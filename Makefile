SHELL := /bin/bash

.PHONY: build
build:
	go build 

.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: coverage
coverage: 
	go tool cover -html=coverage.out
