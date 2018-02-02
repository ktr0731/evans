SHELL := /bin/bash
VERSION := $(shell grep 'version = ' < main.go | sed -r 's/\sversion = "(.*)".*/\1/')

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
test:
	go test -race -v $(shell glide novendor)

.PHONY: coverage
coverage: 
	go tool cover -html=coverage.out

.PHONY: brew-update
brew-update:
	bash .circleci/scripts/entrypoint.bash $(VERSION)
