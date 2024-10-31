BUILD_DIR = bin
PACKAGE = $(shell go list -m)

bin/: ; mkdir -p $@
bin/mockgen: | bin/
	GOBIN="$(realpath $(dir $@))" go install go.uber.org/mock/mockgen@v0.5.0

bin/golangci-lint: | bin/
	GOBIN="$(realpath $(dir $@))" go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3


.PHONY: lint
lint: | bin/golangci-lint ## Run code linters
	@PATH="$(realpath bin):$$PATH" golangci-lint run

.PHONY: test
test: ## Run unit-tests
	@go test ./internal/... -race -v -count=1 ./internal/...


.PHONY: mock
mock: | bin/mockgen ## Generate mocks
	@PATH="$(realpath bin):$$PATH" go generate ./...


.PHONY: build_server
build_server: ## Build a binary server
	CGO_ENABLED=0 go build -o ${BUILD_DIR}/app ${PACKAGE}/cmd/server

.PHONY: build_client
build_client: ## Build a binary client
	CGO_ENABLED=0 go build -o ${BUILD_DIR}/app ${PACKAGE}/cmd/client

.PHONY: startd
startd: ## Compose run detached
	docker compose -f ./build/docker-compose.yml up -d --build --remove-orphans

.PHONY: start
start: ## Compose run
	docker compose -f ./build/docker-compose.yml up --build --remove-orphans

.PHONY: stop
stop: ## Compose down
	docker compose -f ./build/docker-compose.yml down --remove-orphans

.PHONY: log
log: ## Compose logs
	docker compose -f ./build/docker-compose.yml logs -f

.PHONY: logserver
logserver: ## Compose server logs
	docker compose -f ./build/docker-compose.yml logs -f server

.PHONY: logclient
logclient: ## Compose client logs
	docker compose -f ./build/docker-compose.yml logs -f client

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'