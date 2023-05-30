package common

import (
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/utils"
)

// MustGetEnv extract environment variable or abort with an error message
func MustGetEnv(name string) string {
	v, ok := GetEnv(name)
	if !ok {
		log.Fatal().Str("env_var", name).Msg("Missing environment variable")
	}
	return v
}

// MustGetEnvFlag extracts and environment variable and returns a boolean
// corresponding to its value ("true" is true, anything else is false).
// If the environment variable is not found, the program panics with an error message.
func MustGetEnvFlag(name string) bool {
	v, err := utils.GetenvBool(name)
	if err != nil {
		log.Fatal().Str("env_var", name).Err(err).Msg("Failed to determine flag from environment")
	}
	return v
}

// GetEnv attempts to get an environment variable
func GetEnv(name string) (string, bool) {
	return os.LookupEnv(name)
}

// GetEnvOrFallback attempts to get an environment variable or fallback
// to the provided default value.
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
