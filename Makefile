OUTPUT_DIR = ./bin
CHAINCODE_BIN = $(OUTPUT_DIR)/chaincode
ORCHESTRATOR_BIN = $(OUTPUT_DIR)/orchestrator
PROJECT_ROOT = .
protos = $(PROJECT_ROOT)/lib/assets
go_src = $(shell find . -type f -name '*.go')

protobufs = $(wildcard $(protos)/*.proto)
pbgo = $(protobufs:.proto=.pb.go)

all: $(ORCHESTRATOR_BIN) $(CHAINCODE_BIN)

$(ORCHESTRATOR_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR)
	go build -o $(ORCHESTRATOR_BIN) .

$(CHAINCODE_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR)
	go build -o $(CHAINCODE_BIN) ./chaincode

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
proto-codegen: $(pbgo)

.PHONY: clean
clean: clean-protos
	rm -rf $(OUTPUT_DIR)

.PHONY: test
test:
	go test -cover ./...

.PHONY: clean-protos
clean-protos:
	rm $(wildcard $(protos)/*.pb.go)
