.PHONY: help
.DEFAULT_GOAL := help

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

test: ## Run all test
	go test ./... -v -race -cover -count=1

lint: ## Run lint
	golangci-lint run

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
