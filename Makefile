assets = ./lib/assets
protos = ./lib/protos

orchestrator: protos
	go build -o orchestrator .

protos: $(assets)/node/node.pb.go $(assets)/objective/objective.pb.go

$(assets)/node/node.pb.go: $(protos)/node.proto
	protoc --proto_path=lib/protos --go-grpc_out=lib/assets/node --go_out=lib/assets/node lib/protos/node.proto

$(assets)/objective/objective.pb.go: $(protos)/objective.proto
	protoc --proto_path=lib/protos --go-grpc_out=lib/assets/objective --go_out=lib/assets/objective lib/protos/objective.proto

.PHONY: clean
clean:
	rm orchestrator
