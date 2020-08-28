package utils

import (
	"os"
)

func GetServerAddress() string {
	const defaultServerAddress = "127.0.0.1:8000"
	serverAddress, present := os.LookupEnv("SERVER_ADDRESS")
	if present {
		return serverAddress
	}
	return defaultServerAddress
}
