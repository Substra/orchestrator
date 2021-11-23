// chaincode is the hyperledger-fabric chaincode exposing contractapi.
// It relies on the orchestration library for most of the asset handling.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/contracts"
	"github.com/owkin/orchestrator/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const httpPort = "8484"

func main() {
	utils.InitLogging()

	CCID := os.Getenv("CHAINCODE_CCID")

	if metricsEnabled, _ := utils.GetenvBool("METRICS_ENABLED"); metricsEnabled {
		http.Handle("/metrics", promhttp.Handler())

		// Expose HTTP endpoints
		go func() {
			err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil)
			if err != nil {
				log.WithError(err).WithField("port", httpPort).Error("failed to serve HTTP endpoints")
			}
		}()
	}

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
