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

// chaincode is the hyperledger-fabric chaincode exposing contractapi.
// It relies on the orchestration library for most of the asset handling.
package main

import (
	"io/ioutil"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/contracts"
	"github.com/owkin/orchestrator/utils"
)

func main() {
	utils.InitLogging()

	CCID := os.Getenv("CHAINCODE_CCID")

	allContracts := contracts.NewContractCollection().GetAllContracts()

	cc, err := contractapi.NewChaincode(allContracts...)

	if err != nil {
		log.WithError(err).Fatal("error creating substra chaincode")
	}

	key, err := ioutil.ReadFile(os.Getenv("TLS_KEY_FILE"))
	if err != nil {
		log.WithError(err).WithField("path", os.Getenv("TLS_KEY_FILE")).Fatal("unable to read key file")
	}

	cert, err := ioutil.ReadFile(os.Getenv("TLS_CERT_FILE"))
	if err != nil {
		log.WithError(err).WithField("path", os.Getenv("TLS_CERT_FILE")).Fatal("unable to read cert file")
	}

	ca, err := ioutil.ReadFile(os.Getenv("TLS_ROOTCERT_FILE"))
	if err != nil {
		log.WithError(err).WithField("path", os.Getenv("TLS_ROOTCERT_FILE")).Fatal("unable to read CA cert file")
	}

	server := &shim.ChaincodeServer{
		CCID:    CCID,
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
	log.Info("starting substra chaincode server")
	if err = server.Start(); err != nil {
		log.WithError(err).WithField("title", cc.Info.Title).WithField("version", cc.Info.Version).Fatal("error happened while starting chaincode")
	}
}
