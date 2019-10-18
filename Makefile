.DEFAULT_GOAL := help
.PHONY: run build test lint check cover install-linters format cover generate-mocks help

PACKAGES = $(shell find ./src -type d -not -path '\./src')

COMMIT=$$(git rev-parse HEAD)
GOLDFLAGS="-X main.gitCommit=$(COMMIT)"

DATABASE_URL ?= "postgresql://postgres@localhost:54320/wallet?sslmode=disable"

run: ## Run the wallet service. To add arguments, do `make ARGS="--foo" run`.
	go run -ldflags $(GOLDFLAGS) cmd/wallet/wallet.go ${ARGS}

build: ## Build wallet binary
	go build -ldflags $(GOLDFLAGS) cmd/wallet/wallet.go

update-db: ## Updates the database to the latest schema. To change the database URL, do `make DATABASE_URL="..." update-db`.
	migrate -database $(DATABASE_URL) -path ./migrations up

test: ## Run tests
	go test ./... -timeout=1m -cover ${PARALLEL}

test-race: ## Run tests with -race
	go test ./... -timeout=1m -race ${PARALLEL}

lint: ## Run linters. Use make install-linters first.
	golangci-lint run -c .golangci.yml ./...
	@# The govet version in golangci-lint is out of date and has spurious warnings, run it separately
	go vet -all ./...

check: lint test ## Run tests and linters

cover: ## Runs tests on ./src/ with HTML code coverage
	go test -cover -coverprofile=cover.out -coverpkg=github.com/xsleonard/gokit-example/... ./...
	go tool cover -html=cover.out

install-linters: ## Install linters
	# For some reason this install method is not recommended, see https://github.com/golangci/golangci-lint#install
	# However, they suggest `curl ... | bash` which we should not do
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

format: ## Formats the code. Must have goimports installed (use make install-linters).
	goimports -w -local github.com/xsleonard/gokit-example ./...
	# This performs code simplifications
	gofmt -s -w ./...

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
