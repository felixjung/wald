.PHONY: test fmt lint build install download help

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.

.DEFAULT_GOAL := help
help: ## Prints this help
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	@go test ./...

fmt: ## Format code with golangci-lint
	@go tool golangci-lint fmt ./...

lint: ## Run golangci-lint
	@go tool golangci-lint run ./...

build: download ## Build the program
	@mkdir -p bin
	@go build -ldflags "$(LDFLAGS)" -o bin/wald github.com/felixjung/wald/cmd/wald

install: build ## Build and install to /usr/local/bin (override with INSTALL_DIR=/path; may require sudo)
	@set -e; \
	target_dir="$${INSTALL_DIR:-/usr/local/bin}"; \
	mkdir -p "$$target_dir"; \
	if [ ! -w "$$target_dir" ]; then \
		echo "No write permission for $$target_dir. Use 'sudo make install' or set INSTALL_DIR to a writable path."; \
		exit 1; \
	fi; \
	install -m 755 bin/wald "$$target_dir/wald"; \
	echo "Installed bin/wald to $$target_dir/wald"

download: ## Download dependencies
	@echo Download go.mod dependencies
	@go mod download
	@go mod tidy
