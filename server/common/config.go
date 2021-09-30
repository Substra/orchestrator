package common

import (
	"io/ioutil"

	"github.com/go-playground/log/v7"
	"gopkg.in/yaml.v2"
)

// OrchestratorConfiguration describe the business config of the server.
// Contrary to the technical conf item which are received as env var,
// some business details are passed as a yml config file.
type OrchestratorConfiguration struct {
	// map of channels -> organizations
	Channels map[string][]string `yaml:"channels"`
}

const Version = "dev"

func NewConfig(path string) *OrchestratorConfiguration {
	conf := new(OrchestratorConfiguration)

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
