package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SSHUser     string
	FLRApp      string
	FLR_DB      string
	FLR_METRICS string
	FLR_OPC     string
	PASSWORD    string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return Config{
		SSHUser:     os.Getenv("SSH_USER"),
		FLRApp:      os.Getenv("FLR_APP"),
		FLR_DB:      os.Getenv("FLR_DB"),
		FLR_METRICS: os.Getenv("FLR_METRICS"),
		FLR_OPC:     os.Getenv("FLR_OPC"),
		PASSWORD:    os.Getenv("PASSWORD"),
	}
}
