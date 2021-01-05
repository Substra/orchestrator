assets = ./lib/assets
OUTPUT_DIR = ./bin
CHAINCODE_BIN = $(OUTPUT_DIR)/chaincode
ORCHESTRATOR_BIN = $(OUTPUT_DIR)/orchestrator
PROJECT_ROOT = .
go_src = $(shell find . -type f -name '*.go')

protobufs = $(wildcard $(assets)/*/*.proto) $(wildcard $(assets)/*.proto)
pbgo = $(protobufs:.proto=.pb.go)

all: $(ORCHESTRATOR_BIN) $(CHAINCODE_BIN)

$(ORCHESTRATOR_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR)
	go build -o $(ORCHESTRATOR_BIN) .

$(CHAINCODE_BIN): $(pbgo) $(go_src) $(OUTPUT_DIR)
	go build -o $(CHAINCODE_BIN) ./chaincode

$(OUTPUT_DIR):
	mkdir $(OUTPUT_DIR)

$(pbgo): %.pb.go: %.proto
	protoc --proto_path=$(PROJECT_ROOT) --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=$(PROJECT_ROOT) --go_out=$(PROJECT_ROOT) $<

.PHONY: protos
proto-codegen: $(pbgo)

.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)

.PHONY: test
test:
	go test -cover ./...

.PHONY: clean-protos
clean-protos:
	rm $(wildcard $(assets)/*/*.pb.go) $(wildcard $(assets)/*.pb.go)
