// Copyright 2020 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io/ioutil"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/node"
	"github.com/owkin/orchestrator/chaincode/objective"
)

func main() {
	cLog := console.New(true)
	log.AddHandler(cLog, log.AllLevels...)

	nodeContract := node.NewSmartContract()
	nodeContract.Name = "org.substra.node"
	objectiveContract := objective.NewSmartContract()
	objectiveContract.Name = "org.substra.objective"

	cc, err := contractapi.NewChaincode(nodeContract, objectiveContract)

	if err != nil {
		log.Fatal("error creating substra chaincode", err.Error())
	}

	key, err := ioutil.ReadFile(os.Getenv("TLS_KEY_FILE"))
	if err != nil {
		log.Errorf("unable to read key file with path=%s, error: %s", os.Getenv("TLS_KEY_FILE"), err)
	}

	cert, err := ioutil.ReadFile(os.Getenv("TLS_CERT_FILE"))
	if err != nil {
		log.Errorf("unable to read cert file with path %s, error: %s", os.Getenv("TLS_CERT_FILE"), err)
	}

	ca, err := ioutil.ReadFile(os.Getenv("TLS_ROOTCERT_FILE"))
	if err != nil {
		log.Errorf("unable to read ca cert file with path: %s, error: %s", os.Getenv("TLS_CERT_FILE"), err)
	}

	server := &shim.ChaincodeServer{
		CCID:    os.Getenv("CHAINCODE_CCID"),
		Address: os.Getenv("CHAINCODE_ADDRESS"),
		CC:      cc,
		TLSProps: shim.TLSProperties{
			Disabled:      false,
			Key:           key,
			Cert:          cert,
			ClientCACerts: ca,
		},
	}

	// Start the chaincode external server
	log.Infof("starting substra chaincode server")
	if err = server.Start(); err != nil {
		log.Errorf("error happened while starting chaincode %s, version: %s : %s", cc.Info.Title, cc.Info.Version, err)
	}
}
