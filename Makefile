OUTPUT_DIR = ./bin
CHAINCODE_BIN = $(OUTPUT_DIR)/chaincode
ORCHESTRATOR_BIN = $(OUTPUT_DIR)/orchestrator
LIBCODEGEN_BIN = $(OUTPUT_DIR)/libcodegen
FORWARDER_BIN = $(OUTPUT_DIR)/forwarder
PROJECT_ROOT = .
MIGRATIONS_DIR = $(PROJECT_ROOT)/server/standalone/migration
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
orchestrator: $(FORWARDER_BIN)

.PHONY: codegen
codegen: $(pbgo) $(migrations_binpack) $(lib_generated)

.PHONY: lint
lint: codegen
	golangci-lint run

$(ORCHESTRATOR_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(migrations_binpack) $(lib_generated)
	go build -o $(ORCHESTRATOR_BIN) ./server

$(CHAINCODE_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR) $(lib_generated)
	go build -o $(CHAINCODE_BIN) ./chaincode

$(LIBCODEGEN_BIN): $(PROJECT_ROOT)/lib/codegen/main.go
	go build -o $(LIBCODEGEN_BIN) $(PROJECT_ROOT)/lib/codegen

$(FORWARDER_BIN): ${go_src} $(OUTPUT_DIR) $(pbgo)
	go build -o $(FORWARDER_BIN) $(PROJECT_ROOT)/server/distributed/forwarder

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

.PHONY: clean
clean: clean-protos clean-migrations-binpack
	rm -rf $(OUTPUT_DIR)

.PHONY: test
test: codegen
	go test -cover ./... -short

.PHONE: e2e-tests
e2e-tests: codegen
	go test ./e2e -count=1 -tags e2e

.PHONY: clean-protos
clean-protos:
	rm $(wildcard $(protos)/*.pb.go)

.PHONY: clean-migrations-binpack
clean-migrations-binpack:
	rm $(migrations_binpack)
