package utils

import (
	"os"
	"strconv"
)

// GetenvBool returns an environment variable as boolean
func GetenvBool(key string) (bool, error) {
	env := os.Getenv(key)
	b, err := strconv.ParseBool(env)
	if err != nil {
		return false, err
	}
	return b, nil
}
