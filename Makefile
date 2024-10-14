#!/usr/bin/make -f

all: format lint test

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	./contrib/scripts/go-mod-all.sh verify
	./contrib/scripts/go-mod-all.sh tidy
	@echo "--> Download go modules to local cache"
	./contrib/scripts/go-mod-all.sh download

.PHONY: go.sum

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_version=v1.60.3

lint-install:
	@if golangci-lint version --format json | jq -r .version | sed 's/^v//; s/^/v/' | grep $(golangci_version); then \
		echo "golangci-lint $(golangci_version) is already installed"; \
	else \
		echo "Installing golangci-lint $(golangci_version)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version); \
	fi

lint: lint-install
	@echo "--> Running linter"
	golangci-lint run -v --out-format=tab

format: lint-install
	golangci-lint run -v --out-format=tab --fix

###############################################################################
###                                 Tests                                   ###
###############################################################################

test:
	@echo "--> Running tests"
	./contrib/scripts/go-test-all.sh

test-count:
	./contrib/scripts/go-test-all.sh -cpu 1 -count 1 -cover

.PHONY: test
