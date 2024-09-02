package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SSHUser       string
	FLRApp        string
	FLR_DB        string
	FLR_METRICS   string
	FLR_OPC       string
	PASSWORD      string
	SBS_CORE      string
	SBS_BRIGDE    string
	SBS_PUPPET    string
	SBS_INTERFACE string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return Config{
		SSHUser:       os.Getenv("SSH_USER"),
		FLRApp:        os.Getenv("FLR_APP"),
		FLR_DB:        os.Getenv("FLR_DB"),
		FLR_METRICS:   os.Getenv("FLR_METRICS"),
		FLR_OPC:       os.Getenv("FLR_OPC"),
		PASSWORD:      os.Getenv("PASSWORD"),
		SBS_CORE:      os.Getenv("SBS_CORE"),
		SBS_BRIGDE:    os.Getenv("SBS_BRIGDE"),
		SBS_PUPPET:    os.Getenv("SBS_PUPPET"),
		SBS_INTERFACE: os.Getenv("SBS_INTERFACE"),
	}
}
