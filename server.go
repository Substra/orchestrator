package main

import (
	"log"
	"net"

	"github.com/substrafoundation/substra-orchestrator/node"
	"github.com/substrafoundation/substra-orchestrator/objective"
	"google.golang.org/grpc"
)

func main() {
	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	server := grpc.NewServer()
	node.RegisterNodeServiceServer(server, &node.Server{})
	objective.RegisterObjectiveServiceServer(server, &objective.Server{})

	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to server grpc server on port 9000: %v", err)
	}
}
