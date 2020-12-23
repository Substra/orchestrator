assets = ./lib/assets
protos = ./lib/protos
CHAINCODE_BIN = chaincode.bin
ORCHESTRATOR_BIN = orchestrator

all: orchestrator chaincode

orchestrator: protos
	go build -o $(ORCHESTRATOR_BIN) .

chaincode: protos
	go build -o $(CHAINCODE_BIN) ./chaincode

protos: $(assets)/node/node.pb.go $(assets)/objective/objective.pb.go

$(assets)/node/node.pb.go: $(protos)/node.proto
	protoc --proto_path=lib/protos --go-grpc_out=lib/assets/node --go_out=lib/assets/node lib/protos/node.proto

$(assets)/objective/objective.pb.go: $(protos)/objective.proto
	protoc --proto_path=lib/protos --go-grpc_out=lib/assets/objective --go_out=lib/assets/objective lib/protos/objective.proto

.PHONY: clean
clean:
	rm $(ORCHESTRATOR_BIN)
	rm $(CHAINCODE_BIN)

.PHONY: test
test:
	go test -cover ./...
