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
all: chaincode orchestrator forwarder

.PHONY: chaincode
chaincode: $(CHAINCODE_BIN)

.PHONY: orchestrator
orchestrator: $(ORCHESTRATOR_BIN)

.PHONY: forwarder
forwarder: $(FORWARDER_BIN)

.PHONY: codegen
codegen: $(pbgo) $(migrations_binpack) $(lib_generated)

.PHONY: lint
lint: codegen mocks
	golangci-lint run

$(ORCHESTRATOR_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(migrations_binpack) $(lib_generated)
	go build -o $(ORCHESTRATOR_BIN) -ldflags="-X 'server.common.Version=$(VERSION)'" ./server

$(CHAINCODE_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(lib_generated)
	go build -o $(CHAINCODE_BIN) -ldflags="-X 'chaincode.info.Version=$(VERSION)'" ./chaincode

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
proto-codegen: $(pbgo)

.PHONY: mocks
mocks:
	mockery --dir $(PROJECT_ROOT)/lib/event --all --inpackage --quiet
	mockery --dir $(PROJECT_ROOT)/lib/service --all --inpackage --quiet
	mockery --dir $(PROJECT_ROOT)/lib/persistence --all --output $(PROJECT_ROOT)/lib/persistence/mocks --quiet
	mockery --dir $(PROJECT_ROOT)/forwarder/event --all --inpackage --quiet
	mockery --dir $(PROJECT_ROOT)/chaincode --all --output $(PROJECT_ROOT)/chaincode/mocks --quiet

.PHONY: clean
clean: clean-protos clean-migrations-binpack clean-generated clean-mocks
	rm -rf $(OUTPUT_DIR)

.PHONY: test
test: codegen mocks
	go test -race -cover ./... -short -timeout 30s

.PHONY: clean-mocks
clean-mocks:
	-rm $(PROJECT_ROOT)/lib/service/mock_*.go
	-rm $(PROJECT_ROOT)/lib/event/mock_*.go
	-rm -r $(PROJECT_ROOT)/lib/persistence/mocks
	-rm -r $(PROJECT_ROOT)/forwarder/event/mock_*.go
	-rm -r $(PROJECT_ROOT)/chaincode/mocks

.PHONY: clean-protos
clean-protos:
	-rm $(wildcard $(protos)/*.pb.go)

.PHONY: clean-migrations-binpack
clean-migrations-binpack:
	-rm  $(migrations_binpack)

.PHONY: clean-generated
clean-generated:
	-rm $(lib_generated)
