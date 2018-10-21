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
test: vet lint unit-test e2e-test

.PHONY: unit-test
unit-test: deadcode-test
	go test -race $(shell go list ./... | grep -v tests)

.PHONY: e2e-test
e2e-test: deadcode-test
	go test -tags e2e -race ./tests/...

# to find uninitialized dependencies
.PHONY: deadcode-test
deadcode-test:
	gometalinter --vendor --disable-all --enable=deadcode di

.PHONY: vet
vet:
	@gometalinter --vendor --disable-all $(shell go list ./... | grep -v tests)

.PHONY: deadcode
deadcode:
	@gometalinter --vendor --disable-all --enable=deadcode ./...

.PHONY: lint
lint:
	# ignore comments for exported objects
	# ignore Err prefix
	gometalinter --vendor --disable-all --enable=golint --exclude="(should have comment|ErrFoo)" ./...

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
