// chaincode is the hyperledger-fabric chaincode exposing contractapi.
// It relies on the orchestration library for most of the asset handling.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/chaincode/contracts"
	"github.com/substra/orchestrator/utils"
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
				log.Error().Err(err).Str("port", httpPort).Msg("failed to serve HTTP endpoints")
			}
		}()
	}

	allContracts := contracts.NewContractCollection().GetAllContracts()

	cc, err := contractapi.NewChaincode(allContracts...)

	if err != nil {
		log.Fatal().Err(err).Msg("error creating substra chaincode")
	}

	key, err := ioutil.ReadFile(os.Getenv("TLS_KEY_FILE"))
	if err != nil {
		log.Fatal().Err(err).Str("path", os.Getenv("TLS_KEY_FILE")).Msg("unable to read key file")
	}

	cert, err := ioutil.ReadFile(os.Getenv("TLS_CERT_FILE"))
	if err != nil {
		log.Fatal().Err(err).Str("path", os.Getenv("TLS_CERT_FILE")).Msg("unable to read cert file")
	}

	ca, err := ioutil.ReadFile(os.Getenv("TLS_ROOTCERT_FILE"))
	if err != nil {
		log.Fatal().Err(err).Str("path", os.Getenv("TLS_ROOTCERT_FILE")).Msg("unable to read CA cert file")
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
	log.Info().Msg("starting substra chaincode server")
	if err = server.Start(); err != nil {
		log.Fatal().Err(err).Str("title", cc.Info.Title).Str("version", cc.Info.Version).Msg("error happened while starting chaincode")
	}
}
