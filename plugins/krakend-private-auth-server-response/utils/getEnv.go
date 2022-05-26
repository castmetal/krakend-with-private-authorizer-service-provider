package utils

import (
	"os"
)

func GetEnvVariable(key string) string {

	return os.Getenv(key)
}
