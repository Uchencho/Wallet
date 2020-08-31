package utils

import (
	"log"
	"os"
)

func GetServerAddress() string {
	const defaultServerAddress = "127.0.0.1:8000"
	serverAddress, present := os.LookupEnv("SERVER_ADDRESS")
	if present {
		log.Println("Found default address, ", serverAddress)
		return serverAddress
	}
	return defaultServerAddress
}
