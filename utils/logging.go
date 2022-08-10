package utils

import (
	"errors"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/json"
	orcerrors "github.com/substra/orchestrator/lib/errors"
)

// InitLogging configure log library to output to console with appropriate levels
func InitLogging() {
	handler := json.New(os.Stdout)

	levels := getLevelsFromEnv()

	log.AddHandler(handler, levels...)

	log.SetWithErrorFn(handleOrcError)
}

// GetLogLevelFromEnv gets the logging level from the environment variable LOG_LEVEL
func GetLogLevelFromEnv() log.Level {
	level := os.Getenv("LOG_LEVEL")
	return parseLevel(level)
}

// handleOrcError augment the log output with error's source
func handleOrcError(entry log.Entry, err error) log.Entry {
	out := entry.WithField("error", err.Error())

	orcError := new(orcerrors.OrcError)
	if errors.As(err, &orcError) {
		out = out.WithField("source", orcError.Source())
	}

	return out
}

// getLevelsFromEnv set logging level to match the level provided by GetLogLevelFromEnv.
// It defaults to INFO if the env var does not exist.
func getLevelsFromEnv() []log.Level {
	minLevel := GetLogLevelFromEnv()

	levels := make([]log.Level, 0)

	for _, level := range log.AllLevels {
		if level >= minLevel {
			levels = append(levels, level)
		}
	}

	return levels
}

// parseLevel attempts to match a string to its log.Level.
// Defaults to log.InfoLevel if no match found.
func parseLevel(s string) log.Level {
	switch s {
	case "DEBUG":
		return log.DebugLevel
	case "INFO":
		return log.InfoLevel
	case "NOTICE":
		return log.NoticeLevel
	case "WARN":
		return log.WarnLevel
	case "ERROR":
		return log.ErrorLevel
	case "PANIC":
		return log.PanicLevel
	case "ALERT":
		return log.AlertLevel
	case "FATAL":
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}
