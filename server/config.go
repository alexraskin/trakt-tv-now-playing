package server

import (
	"log"
	"os"
	"strings"
)

func GetTraktAPIKey() string {
	envValue := os.Getenv("TRAKT_API_KEY")

	if strings.HasPrefix(envValue, "/") {
		data, err := os.ReadFile(envValue)
		if err != nil {
			log.Fatalf("Failed to read Trakt API key from secret file: %v", err)
		}
		return strings.TrimSpace(string(data))
	}

	if envValue == "" {
		log.Fatal("TRAKT_API_KEY is not set")
	}

	return envValue
}
