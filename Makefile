OUTPUT_DIR = ./bin
CHAINCODE_BIN = $(OUTPUT_DIR)/chaincode
ORCHESTRATOR_BIN = $(OUTPUT_DIR)/orchestrator
PROJECT_ROOT = .
MIGRATIONS_DIR = $(PROJECT_ROOT)/server/standalone/migration
VERSION = dirty-$(shell git rev-parse --short HEAD)
protos = $(PROJECT_ROOT)/lib/asset
go_src = $(shell find . -type f -name '*.go')
files_to_lint := $(shell gofmt -l -s .)
sql_migrations = $(wildcard $(MIGRATIONS_DIR)/*.sql)

protobufs = $(wildcard $(protos)/*.proto)
pbgo = $(protobufs:.proto=.pb.go)

# Disable cgo since we don't use it and linking is broken with some version of go1.18 on macos
build_env = CGO_ENABLED=0

.PHONY: all
all: chaincode orchestrator  ## Build all binaries

.PHONY: chaincode
chaincode: $(CHAINCODE_BIN)  ## Build chaincode binary

.PHONY: orchestrator
orchestrator: $(ORCHESTRATOR_BIN)  ## Build server binary

.PHONY: codegen
codegen: $(pbgo) $(migrations_binpack)  ## Build codegen tool

.PHONY: gofmt
gofmt: gofmt -l -s .

.PHONY: lint-gofmt ## Lint the codebase with gofmt
lint-gofmt:
	@if [ "$(files_to_lint)" ]; then \
		@echo "Following files should be linted:"; \
		echo "$(files_to_lint)"; \
		exit 1; \
	fi
.PHONY: lint
lint: codegen mocks lint-gofmt  ## Analyze the codebase
	golangci-lint run

.PHONY: format
format: codegen # Format codebase
	gofmt -s -w .

$(ORCHESTRATOR_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(lib_generated)
	$(build_env) go build -o $(ORCHESTRATOR_BIN) -ldflags="-X 'github.com/substra/orchestrator/server/common.Version=$(VERSION)'" ./server

$(CHAINCODE_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(lib_generated)
	$(build_env) go build -o $(CHAINCODE_BIN) -ldflags="-X 'github.com/substra/orchestrator/chaincode/info.Version=$(VERSION)'" ./chaincode

$(OUTPUT_DIR):
	mkdir $(OUTPUT_DIR)

$(pbgo): %.pb.go: %.proto
	protoc --proto_path=$(protos) \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	--go-grpc_out=$(protos) \
	--go_out=$(protos) \
	$<

.PHONY: proto-codegen
proto-codegen: $(pbgo)  ## Generate go code from proto files

.PHONY: mocks
mocks:  ## Generate mocks for public interfaces
	mockery --dir $(PROJECT_ROOT) --all --inpackage --quiet

.PHONY: clean
clean: clean-protos clean-mocks  ## Remove all generated code
	rm -rf $(OUTPUT_DIR)

.PHONY: test
test: codegen mocks  ## Run unit-tests
	go test -race -cover ./... -short -timeout 30s

.PHONY: clean-mocks
clean-mocks:  ## Remove generated mocks
	find $(PROJECT_ROOT) -name "mock_*.go" -delete

.PHONY: clean-protos
clean-protos:  ## Remove go code generated from proto files
	-rm $(wildcard $(protos)/*.pb.go)

.PHONY: docs-charts
docs-charts: ## Generate Helm chart documentation
	$(MAKE) -C charts doc

### Makefile

.PHONY: help
help:  ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
