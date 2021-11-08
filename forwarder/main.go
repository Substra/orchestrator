// chaincode event forwarder.
// This binary listens to chaincode events for multiple channels and forward them to an AMQP exchange.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/forwarder/event"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"github.com/owkin/orchestrator/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

// mustGetEnv extract environment variable or abort with an error message
// Every env var is prefixed by FORWARDER_
func mustGetEnv(name string) string {
	n := "FORWARDER_" + name
	v, ok := os.LookupEnv(n)
	if !ok {
		log.WithField("env_var", n).Fatal("Missing environment variable")
	}
	return v
}

type forwarderConf struct {
	// map of channel -> chaincodes
	Listeners map[string][]string `yaml:"listeners"`
}

func main() {
	utils.InitLogging()

	networkConfig := mustGetEnv("NETWORK_CONFIG")
	indexFile := mustGetEnv("EVENT_INDEX_FILE")

	eventIdx, err := event.NewIndex(indexFile)
	if err != nil {
		log.WithError(err).WithField("indexFile", indexFile).Fatal("cannot instanciate event indexer")
	}

	rabbitDSN := mustGetEnv("AMQP_DSN")
	session := common.NewSession("orchestrator", rabbitDSN)
	defer session.Close()

	wallet := wallet.New(mustGetEnv("FABRIC_CERT"), mustGetEnv("FABRIC_KEY"))

	config := config.FromFile(networkConfig)
	log.Info("network config loaded")

	mspID := mustGetEnv("MSPID")

	conf := getConf(mustGetEnv("CONFIG_PATH"))

	wg := new(sync.WaitGroup)

	for channel, chaincodes := range conf.Listeners {
		forwarder := event.NewForwarder(channel, session)
		log.WithField("channel", channel).Info("instanciated AMQP forwarder")

		for _, chaincode := range chaincodes {
			ccData := event.ListenerChaincodeData{
				Wallet:    wallet,
				Config:    config,
				MSPID:     mspID,
				Channel:   channel,
				Chaincode: chaincode,
			}
			wg.Add(1)
			go listenToChannel(ccData, forwarder, eventIdx, wg)
		}
	}

	wg.Wait()
	log.Debug("all listeners ready")

	http.HandleFunc("/", healthcheck)
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.WithError(err).Fatal("Could not spawn http server")
	}
}

// Pretty basic liveness
func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func listenToChannel(ccData event.ListenerChaincodeData, forwarder *event.Forwarder, eventIdx event.Indexer, wg *sync.WaitGroup) {
	var listener *event.Listener

	for {
		var err error
		listener, err = event.NewListener(&ccData, eventIdx, forwarder.Forward)

		if err != nil {
			waitTime := time.Second * 5
			log.WithError(err).WithField("waitTime", waitTime).Error("Failed to instanciate listener, retrying")
			time.Sleep(waitTime)
			continue
		}

		break
	}

	wg.Done()

	defer listener.Close()
	log.WithField("channel", ccData.Channel).WithField("chaincode", ccData.Chaincode).Info("Listening to channel events")

	listener.Listen()
}

func getConf(path string) *forwarderConf {
	conf := new(forwarderConf)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithError(err).Fatal("Failed to read config file")
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse config file")
	}

	return conf
}
