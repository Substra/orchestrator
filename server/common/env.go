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
