package utils

import (
	"errors"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

// InitLogging configure log library to output to console with appropriate levels
func InitLogging() {
	cLog := console.New(true)

	_, noColor := os.LookupEnv("NO_COLOR")
	cLog.SetDisplayColor(!noColor)

	levels := getLevelsFromEnv()

	log.AddHandler(cLog, levels...)

	log.SetWithErrorFn(handleOrcError)
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

// getLevelsFromEnv set logging level to match the LOG_LEVEL environment var.
// It defaults to INFO if the env var does not exist.
func getLevelsFromEnv() []log.Level {
	level := os.Getenv("LOG_LEVEL")
	minLevel := parseLevel(level)

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
