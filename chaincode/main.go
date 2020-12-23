package main

import (
	"log"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/sirupsen/logrus"
	"github.com/substrafoundation/substra-orchestrator/chaincode/ledger"
	"github.com/substrafoundation/substra-orchestrator/chaincode/node"
	nodeService "github.com/substrafoundation/substra-orchestrator/lib/assets/node"
)

var logger = logrus.New()

func main() {

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	nodeServer := nodeService.NewServer(ledger.GetLedgerFromContext)
	chaincode, err := contractapi.NewChaincode(node.NewNodeContract(nodeServer))

	if err != nil {
		log.Fatal("Error create substra chaincode", err.Error())
	}

	if err := chaincode.Start(); err != nil {
		panic(err.Error())
	}
}
