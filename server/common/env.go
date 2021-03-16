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

// server binary exposing a gRPC interface to manage distributed learning asset.
// It can run in either standalone or distributed mode.
// In standalone mode it handle all the logic while in distributed mode everything is delegated to a chaincode.
package common

import (
	"os"

	"github.com/go-playground/log/v7"
)

// envPrefix is the string prefixing environment variables related to the orchestrator
const envPrefix = "ORCHESTRATOR_"

// MustGetEnv extract environment variable or abort with an error message
// Every env var is prefixed with ORCHESTRATOR_
func MustGetEnv(name string) string {
	v, ok := GetEnv(name)
	if !ok {
		log.WithField("env_var", envPrefix+name).Fatal("Missing environment variable")
	}
	return v
}

// MustGetEnvFlag extracts and environment variable and returns a boolean
// corresponding to its value ("true" is true, anything else is false).
// If the environment variable is not found, the program panics with an error message.
// Every env var is prefixed with "ORCHESTRATOR_".
func MustGetEnvFlag(name string) bool {
	return MustGetEnv(name) == "true"
}

// GetEnv attempts to get an environment variable
// Every env var is prefixed by ORCHESTRATOR_
func GetEnv(name string) (string, bool) {
	n := envPrefix + name
	return os.LookupEnv(n)
}
