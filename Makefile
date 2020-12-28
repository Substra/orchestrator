assets = ./lib/assets
CHAINCODE_BIN = chaincode.bin
ORCHESTRATOR_BIN = orchestrator

protobufs = $(wildcard $(assets)/*/*.proto)
pbgo = $(protobufs:.proto=.pb.go)

all: orchestrator chaincode

orchestrator: $(pbgo)
	go build -o $(ORCHESTRATOR_BIN) .

chaincode: $(pbgo)
	go build -o $(CHAINCODE_BIN) ./chaincode

$(pbgo): %.pb.go: %.proto
	protoc --proto_path=$(dir $<) --go-grpc_out=$(dir $@) --go_out=$(dir $@) $<

.PHONY: clean
clean:
	rm $(ORCHESTRATOR_BIN)
	rm $(CHAINCODE_BIN)

.PHONY: test
test:
	go test -cover ./...
