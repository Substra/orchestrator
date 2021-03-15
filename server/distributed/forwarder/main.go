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

// chaincode event forwarder.
// This binary listens to chaincode events for multiple channels and forward them to an AMQP exchange.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/distributed/event"
	"github.com/owkin/orchestrator/server/distributed/wallet"
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
	common.InitLogging()

	networkConfig := mustGetEnv("NETWORK_CONFIG")

	rabbitDSN := mustGetEnv("AMQP_DSN")
	session := common.NewSession("orchestrator", rabbitDSN)
	defer session.Close()

	wallet := wallet.New(mustGetEnv("CERT"), mustGetEnv("KEY"))

	config := config.FromFile(networkConfig)
	log.Info("network config loaded")

	forwarder := event.NewForwarder(session)
	log.Info("instanciated AMQP forwarder")

	mspID := mustGetEnv("MSPID")

	conf := getConf(mustGetEnv("CONFIG_PATH"))

	for channel, chaincodes := range conf.Listeners {
		for _, chaincode := range chaincodes {
			go listenToChannel(wallet, config, forwarder, mspID, chaincode, channel)
		}
	}

	http.HandleFunc("/", healthcheck)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.WithError(err).Fatal("Could not spawn http server")
	}
}

// Pretty basic liveness
func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func listenToChannel(wallet *wallet.Wallet, config core.ConfigProvider, forwarder *event.Forwarder, mspID string, chaincode string, channel string) {
	listener, err := event.NewListener(wallet, config, mspID, channel, chaincode, forwarder.Forward)

	if err != nil {
		log.WithError(err).Fatal("Failed to instanciate listener")
	}

	defer listener.Close()
	log.WithField("channel", channel).WithField("chaincode", chaincode).Info("Listening to channel events")

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
