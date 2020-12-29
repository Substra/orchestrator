assets = ./lib/assets
CHAINCODE_BIN = chaincode.bin
ORCHESTRATOR_BIN = orchestrator
PROJECT_ROOT = .
go_src = $(shell find . -type f -name '*.go')

protobufs = $(wildcard $(assets)/*/*.proto) $(wildcard $(assets)/*.proto)
pbgo = $(protobufs:.proto=.pb.go)

all: $(ORCHESTRATOR_BIN) $(CHAINCODE_BIN)

$(ORCHESTRATOR_BIN): $(pbgo) $(go_src)
	go build -o $(ORCHESTRATOR_BIN) .

$(CHAINCODE_BIN): $(pbgo) $(go_src)
	go build -o $(CHAINCODE_BIN) ./chaincode

$(pbgo): %.pb.go: %.proto
	protoc --proto_path=$(PROJECT_ROOT) --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=$(PROJECT_ROOT) --go_out=$(PROJECT_ROOT) $<

.PHONY: protos
proto-codegen: $(pbgo)

.PHONY: clean
clean:
	rm $(ORCHESTRATOR_BIN)
	rm $(CHAINCODE_BIN)

.PHONY: test
test:
	go test -cover ./...

.PHONY: clean-protos
clean-protos:
	rm $(wildcard $(assets)/*/*.pb.go) $(wildcard $(assets)/*.pb.go)
