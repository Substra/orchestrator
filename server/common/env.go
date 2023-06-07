package common

import (
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/utils"
)

// envPrefix is the string prefixing environment variables related to the orchestrator
const envPrefix = "ORCHESTRATOR_"

// MustGetEnv extract environment variable or abort with an error message
// Every env var is prefixed with ORCHESTRATOR_
func MustGetEnv(name string) string {
	v, ok := GetEnv(name)
	if !ok {
		log.Fatal().Str("env_var", envPrefix+name).Msg("Missing environment variable")
	}
	return v
}

// MustGetEnvFlag extracts and environment variable and returns a boolean
// corresponding to its value ("true" is true, anything else is false).
// If the environment variable is not found, the program panics with an error message.
// Every env var is prefixed with "ORCHESTRATOR_".
func MustGetEnvFlag(name string) bool {
	n := envPrefix + name
	v, err := utils.GetenvBool(n)
	if err != nil {
		log.Fatal().Str("env_var", envPrefix+name).Err(err).Msg("Failed to determine flag from environment")
	}
	return v
}

// GetEnv attempts to get an environment variable
// Every env var is prefixed by ORCHESTRATOR_
func GetEnv(name string) (string, bool) {
	n := envPrefix + name
	return os.LookupEnv(n)
}

// GetEnvOrFallback attempts to get an environment variable or fallback
// to the provided default value.
// Every env var is prefixed by ORCHESTRATOR_.
func GetEnvOrFallback(name string, fallback string) string {
	value, ok := GetEnv(name)
	if !ok {
		value = fallback
	}
	return value
}

// MustParseDuration parse input as a duration or log and exit.
func MustParseDuration(duration string) time.Duration {
	res, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatal().Str("duration", duration).Msg("Cannot parse duration")
	}
	return res
}

// MustParseInt parse input as int or log and exit.
func MustParseInt(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal().Str("input", s).Msg("Cannot parse integer")
	}
	return res
}

// MustParseBool parse input as bool or log and exit
func MustParseBool(s string) bool {
	res, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatal().Str("input", s).Msg("Cannot parse boolean")
	}
	return res
}
