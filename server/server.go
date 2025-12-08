package server

import (
	"os"
)

const (
	DefaultPort = "7540"
	WebDir      = "./web"
)

func GetPort() string {
	if port := os.Getenv("TODO_PORT"); port != "" {
		return port
	}
	return DefaultPort
}
