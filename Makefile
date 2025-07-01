# TinkoffGo Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Binary names
BINARY_NAME=tinkoff-go

# Build targets
.PHONY: all build clean test deps fmt vet examples help proto proto-clean proto-update \
        example-connect example-accounts example-streaming example-real-api \
        run-connect run-accounts run-streaming run-real-api \
        dev-setup lint docker-build docker-run release

all: deps proto fmt vet test build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format Go code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

test: ## Run tests
	$(GOTEST) -v ./...

build: ## Build the main binary
	$(GOBUILD) -o bin/$(BINARY_NAME) -v .

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf bin/

# Proto generation
proto: ## Generate gRPC code from proto files
	@echo "Generating Go code from proto files..."
	cd proto && protoc --go_out=. --go-grpc_out=. *.proto
	@echo "Proto generation complete!"

proto-clean: ## Clean generated proto files
	rm -f proto/*.pb.go

proto-update: ## Update proto files from Tinkoff repository
	@echo "Updating proto files from Tinkoff Invest API..."
	cd proto && rm -rf temp_proto
	cd proto && git clone https://github.com/RussianInvestments/investAPI.git temp_proto
	cd proto && cp -r temp_proto/src/docs/contracts/* .
	cd proto && rm -rf temp_proto
	@echo "Proto files updated. Run 'make proto' to regenerate Go code."

# Example targets
examples: example-connect example-accounts example-streaming example-real-api ## Build all examples

example-connect: ## Build connection example
	$(GOBUILD) -o bin/example-connect ./examples/connect
	@echo "Built example-connect. Run with: ./bin/example-connect"

example-accounts: ## Build accounts example
	$(GOBUILD) -o bin/example-accounts ./examples/accounts
	@echo "Built example-accounts. Run with: ./bin/example-accounts"

example-streaming: ## Build streaming example
	$(GOBUILD) -o bin/example-streaming ./examples/streaming
	@echo "Built example-streaming. Run with: ./bin/example-streaming"

example-real-api: ## Build real API example
	$(GOBUILD) -o bin/example-real-api ./examples/real_api
	@echo "Built example-real-api. Run with: TINKOFF_TOKEN=your_token ./bin/example-real-api"

run-connect: example-connect ## Run connection example
	./bin/example-connect

run-accounts: example-accounts ## Run accounts example
	./bin/example-accounts

run-streaming: example-streaming ## Run streaming example
	./bin/example-streaming

run-real-api: example-real-api ## Run real API example (requires TINKOFF_TOKEN)
	@if [ -z "$(TINKOFF_TOKEN)" ]; then \
		echo "Error: TINKOFF_TOKEN environment variable is required"; \
		echo "Usage: make run-real-api TINKOFF_TOKEN=your_token"; \
		exit 1; \
	fi
	TINKOFF_TOKEN=$(TINKOFF_TOKEN) ./bin/example-real-api

# Development targets
dev-setup: ## Set up development environment
	$(GOGET) -u golang.org/x/tools/cmd/goimports
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

lint: ## Run golangci-lint
	golangci-lint run

# Docker targets
docker-build: ## Build Docker image
	docker build -t tinkoff-go .

docker-run: ## Run in Docker container
	docker run -it --rm tinkoff-go

# Release targets
release: ## Create a release build
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-amd64 -v .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-amd64 -v .
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-windows-amd64.exe -v .
