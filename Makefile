.PHONY: help build clean run tidy

.DEFAULT_GOAL := help

BINARY=./bin/screenpen-go
MAIN=./cmd/screenpen-go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  %-10s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: tidy ## Build the application binary
	go build -o $(BINARY) $(MAIN)

clean: ## Remove built binary
	rm -f $(BINARY)

run: build ## Build and run the application
	$(BINARY)

tidy: ## Tidy Go module dependencies
	go mod tidy
