OUTPUT_DIR = ./bin
CHAINCODE_BIN = $(OUTPUT_DIR)/chaincode
ORCHESTRATOR_BIN = $(OUTPUT_DIR)/orchestrator
LIBCODEGEN_BIN = $(OUTPUT_DIR)/libcodegen
FORWARDER_BIN = $(OUTPUT_DIR)/forwarder
E2E_BIN = $(OUTPUT_DIR)/e2e-tests
PROJECT_ROOT = .
MIGRATIONS_DIR = $(PROJECT_ROOT)/server/standalone/migration
VERSION = dirty-$(shell git rev-parse --short HEAD)
protos = $(PROJECT_ROOT)/lib/asset
go_src = $(shell find . -type f -name '*.go')
sql_migrations = $(wildcard $(MIGRATIONS_DIR)/*.sql)
migrations_binpack = $(MIGRATIONS_DIR)/bindata.go
lib_generated = $(PROJECT_ROOT)/lib/asset/json.go

protobufs = $(wildcard $(protos)/*.proto)
pbgo = $(protobufs:.proto=.pb.go)

.PHONY: all
all: chaincode orchestrator forwarder  ## Build all binaries

.PHONY: chaincode
chaincode: $(CHAINCODE_BIN)  ## Build chaincode binary

.PHONY: orchestrator
orchestrator: $(ORCHESTRATOR_BIN)  ## Build server binary

.PHONY: forwarder
forwarder: $(FORWARDER_BIN)  ## Build event-forwarded binary

.PHONY: codegen
codegen: $(pbgo) $(migrations_binpack) $(lib_generated)  ## Build codegen tool

.PHONY: lint
lint: codegen mocks  ## Analyze the codebase
	golangci-lint run

$(ORCHESTRATOR_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(migrations_binpack) $(lib_generated)
	go build -o $(ORCHESTRATOR_BIN) -ldflags="-X 'github.com/owkin/orchestrator/server/common.Version=$(VERSION)'" ./server

$(CHAINCODE_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(lib_generated)
	go build -o $(CHAINCODE_BIN) -ldflags="-X 'github.com/owkin/orchestrator/chaincode/info.Version=$(VERSION)'" ./chaincode

$(LIBCODEGEN_BIN): $(PROJECT_ROOT)/lib/codegen/main.go
	go build -o $(LIBCODEGEN_BIN) $(PROJECT_ROOT)/lib/codegen

$(FORWARDER_BIN): ${go_src} $(OUTPUT_DIR) $(pbgo) $(lib_generated)
	go build -o $(FORWARDER_BIN) $(PROJECT_ROOT)/forwarder

$(E2E_BIN): $(go_src) $(OUTPUT_DIR) $(pbgo)
	go build -o $(E2E_BIN) $(PROJECT_ROOT)/e2e

$(OUTPUT_DIR):
	mkdir $(OUTPUT_DIR)

$(pbgo): %.pb.go: %.proto
	protoc --proto_path=$(protos) \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	--go-grpc_out=$(protos) \
	--go_out=$(protos) \
	$<

$(migrations_binpack): $(sql_migrations)
	go-bindata -pkg migration -prefix $(MIGRATIONS_DIR) -o $(migrations_binpack) $(MIGRATIONS_DIR)

$(lib_generated): $(LIBCODEGEN_BIN) $(pbgo)
	$(LIBCODEGEN_BIN) -path $(protos) > $(lib_generated)

.PHONY: proto-codegen
proto-codegen: $(pbgo)  ## Generate go code from proto files

.PHONY: mocks
mocks:  ## Generate mocks for public interfaces
	mockery --dir $(PROJECT_ROOT) --all --inpackage --quiet

.PHONY: clean
clean: clean-protos clean-migrations-binpack clean-generated clean-mocks  ## Remove all generated code
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

.PHONY: clean-migrations-binpack
clean-migrations-binpack:  ## Remove generated migration file
	-rm  $(migrations_binpack)

.PHONY: clean-generated
clean-generated:  ## Remove codegen tool
	-rm $(lib_generated)

.PHONY: docs-charts
docs-charts: ## Generate Helm chart documentation
	$(MAKE) -C charts doc

### Makefile

.PHONY: help
help:  ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
