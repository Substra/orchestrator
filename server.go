package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/substrafoundation/substra-orchestrator/lib/node"
	"github.com/substrafoundation/substra-orchestrator/lib/objective"
	"google.golang.org/grpc"
)

// RunServerWithChainCode is exported
func RunServerWithChainCode() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatal("failed guess where")
	}

	if !wallet.Exists("appClient") {
		cert, err := ioutil.ReadFile("/Users/inal/fabric/sampleconfig/msp/signcerts/peer.pem")
		if err != nil {
			log.Fatal("failed guess where")
		}

		key, err := ioutil.ReadFile("/Users/inal/fabric/sampleconfig/msp/keystore/key.pem")
		if err != nil {
			log.Fatal("failed guess where")
		}

		identity := gateway.NewX509Identity("SampleOrg.member", string(cert), string(key))

		wallet.Put("appClient", identity)
	}

	// get config path

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile("/Users/inal/fabric/sampleconfig/configtx.yaml")),
		gateway.WithIdentity(wallet, "appClient"),
	)

	if err != nil {
		log.Fatalf("failed guess where %v", err)
	}

	defer gw.Close()

	network, err := gw.GetNetwork("ch1")
	if err != nil {
		log.Fatal("failed guess where")
	}

	contract := network.GetContract("mycc")
	result, err := contract.SubmitTransaction("RegisterNode", "1")
	if err != nil {
		log.Fatal("failed guess where")
	}

	log.Println(result)
}

// RunServerWithoutChaincode is exported
func RunServerWithoutChaincode() {
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

func main() {
	RunServerWithChainCode()
}
