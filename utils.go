package quickcrdb

import (
	"fmt"
	"os"
	"strconv"
)

func getEnvOrDefault(env, defaultVal string) string {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		return e
	}
}

func getEnvOrDefaultInt(env string, defaultVal int64) int64 {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		intVal, err := strconv.ParseInt(e, 10, 64)
		if err != nil {
			logger.Error().Msg(fmt.Sprintf("Failed to parse string to int '%s'", env))
			os.Exit(1)
		}

		return intVal
	}
}
