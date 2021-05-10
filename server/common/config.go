// Copyright 2021 Owkin Inc.
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
