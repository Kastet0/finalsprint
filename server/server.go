package server

import (
	"os"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

func getPort() string {
	if port := os.Getenv("TODO_PORT"); port != "" {
		return port
	}
	return defaultPort
}
