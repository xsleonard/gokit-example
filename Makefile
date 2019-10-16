.DEFAULT_GOAL := help
.PHONY: run build test lint check cover install-linters format cover help

PACKAGES = $(shell find ./src -type d -not -path '\./src')

COMMIT=$$(git rev-parse HEAD)
GOLDFLAGS="-X main.gitCommit=$(COMMIT)"

run: ## Run the wallet service. To add arguments, do `make ARGS="--foo" run`.
	go run -ldflags $(GOLDFLAGS) cmd/wallet/wallet.go ${ARGS}

build: ## Build wallet binary
	go build -ldflags $(GOLDFLAGS) cmd/wallet/wallet.go

test: ## Run tests
	go test ./... -timeout=1m -cover ${PARALLEL}

test-race: ## Run tests with -race
	go test ./... -timeout=1m -race ${PARALLEL}

lint: ## Run linters. Use make install-linters first.
	gometalinter --deadline=3m -j 2 --disable-all --tests --vendor \
		-E deadcode \
		-E errcheck \
		-E gas \
		-E goconst \
		-E gofmt \
		-E goimports \
		-E golint \
		-E ineffassign \
		-E interfacer \
		-E maligned \
		-E megacheck \
		-E misspell \
		-E nakedret \
		-E structcheck \
		-E unconvert \
		-E unparam \
		-E varcheck \
		-E vet \
		./...

check: lint test ## Run tests and linters

cover: ## Runs tests on ./src/ with HTML code coverage
	go test -cover -coverprofile=cover.out -coverpkg=github.com/xsleonard/gokit-example/... ./...
	go tool cover -html=cover.out

install-linters: ## Install linters
	go get -u github.com/alecthomas/gometalinter
	gometalinter --vendored-linters --install

format: ## Formats the code. Must have goimports installed (use make install-linters).
	goimports -w -local github.com/xsleonard/gokit-example ./...
	# This performs code simplifications
	gofmt -s -w ./...

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
