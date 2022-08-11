package utils

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogging configure log library to output to console with appropriate levels
func InitLogging() {
	level, err := GetLogLevelFromEnv()
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse log level, defaulting to INFO")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
}

// GetLogLevelFromEnv gets the logging level from the environment variable LOG_LEVEL
func GetLogLevelFromEnv() (zerolog.Level, error) {
	level := os.Getenv("LOG_LEVEL")
	level = strings.ToLower(level)
	return zerolog.ParseLevel(level)
}
