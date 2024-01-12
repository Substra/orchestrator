package common

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// OrchestratorConfiguration describe the business config of the server.
// Contrary to the technical conf item which are received as env var,
// some business details are passed as a yml config file.
type OrchestratorConfiguration struct {
	// map of channels -> organizations
	Channels map[string][]string `yaml:"channels"`
}

// Version represents the version of the server, the value is changed at build time
var Version = "dev"

func NewConfig(path string) *OrchestratorConfiguration {
	conf := new(OrchestratorConfiguration)

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read config file")
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}

	return conf
}
