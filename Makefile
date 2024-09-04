export

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Run docker compose
		docker compose up --build -d
.PHONY: compose-up

compose-down: ### Shut down docker compose
		docker compose down --remove-orphans
.PHONY: compose-down

linter-golangci: ### Check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

swag-v1: ### swag init
	swag init -g internal/controller/http/endpoints.go
.PHONY: swag-v1

test: ### Run tests
	go test -v -cover -race ./internal/...
.PHONY: test
