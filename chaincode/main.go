package main

import (
	"log"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/sirupsen/logrus"
	"github.com/substrafoundation/substra-orchestrator/chaincode/node"
	"github.com/substrafoundation/substra-orchestrator/chaincode/objective"
)

var logger = logrus.New()

func main() {

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	nodeContract := node.NewSmartContract()
	nodeContract.Name = "org.substra.node"
	objectiveContract := objective.NewSmartContract()
	objectiveContract.Name = "org.substra.objective"

	chaincode, err := contractapi.NewChaincode(nodeContract, objectiveContract)

	if err != nil {
		log.Fatal("Error creating substra chaincode", err.Error())
	}

	if err := chaincode.Start(); err != nil {
		panic(err.Error())
	}
}
