SHELL := /bin/bash

.PHONY: default clean build fmt lint vet cyclo ineffassign shellcheck errcheck goconst gosec abcgo json-check openapi-check style run test test-postgres cover integration_tests rest_api_tests sqlite_db license before_commit help godoc install_docgo install_addlicense

SOURCES:=$(shell find . -name '*.go')
BINARY:=insights-results-aggregator-exporter
DOCFILES:=$(addprefix docs/packages/, $(addsuffix .html, $(basename ${SOURCES})))

default: build

clean: ## Run go clean
	@go clean

build: ${BINARY} ## Build binary containing service executable

build-cover:	${SOURCES}  ## Build binary with code coverage detection support
	./build.sh -cover

${BINARY}: ${SOURCES}
	./build.sh


install_golangci-lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

fmt: install_golangci-lint ## Run go formatting
	@echo "Running go formatting"
	golangci-lint fmt

lint: install_golangci-lint ## Run go liting
	@echo "Running go linting"
	golangci-lint run --fix

shellcheck: ## Run shellcheck
	./shellcheck.sh

abcgo: ## Run ABC metrics checker
	@echo "Run ABC metrics checker"
	./abcgo.sh ${VERBOSE}

style: fmt lint shellcheck abcgo ## Run all the formatting related commands (fmt, lint, abc) + check shell scripts

run: ${BINARY} ## Build the project and executes the binary
	./$^

test: ${BINARY} ## Run the unit tests
	./unit-tests.sh

cover: test
	@go tool cover -html=coverage.out

coverage:
	@go tool cover -func=coverage.out

license: install_addlicense
	addlicense -c "Red Hat, Inc" -l "apache" -v ./

before_commit: style test test-postgres integration_tests license ## Checks done before commit
	./check_coverage.sh

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

function_list: ${BINARY} ## List all functions in generated binary file
	go tool objdump ${BINARY} | grep ^TEXT | sed "s/^TEXT\s//g"

docs/packages/%.html: %.go
	mkdir -p $(dir $@)
	docgo -outdir $(dir $@) $^
	addlicense -c "Red Hat, Inc" -l "apache" -v $@

godoc: export GO111MODULE=off
godoc: install_docgo install_addlicense ${DOCFILES}

install_docgo: export GO111MODULE=off
install_docgo:
	[[ `command -v docgo` ]] || GO111MODULE=off go get -u github.com/dhconnelly/docgo

install_addlicense: export GO111MODULE=off
install_addlicense:
	[[ `command -v addlicense` ]] || GO111MODULE=off go get -u github.com/google/addlicense
