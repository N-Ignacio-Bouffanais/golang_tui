package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SSHUser           string
	SSH_SERVICE       string
	FLRApp            string
	FLR_DB            string
	FLR_METRICS       string
	FLR_OPC           string
	FLR_FM            string
	PASSWORD          string
	SBS_PASSWORD      string
	SBS_PASS2         string
	SBS_CORE          string
	SBS_BRIGDE        string
	SBS_PUPPET        string
	SBS_INTERFACE     string
	SBS_PLATFORM_API  string
	SBS_PLATFORM_CORE string
	SBS_PLATFORM_DB   string
	SBS_METRICS       string
	SBS_STAGING       string
	SBS_OPC           string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return Config{
		SSHUser:           os.Getenv("SSH_USER"),
		SSH_SERVICE:       os.Getenv("SSH_SERVICE"),
		FLRApp:            os.Getenv("FLR_APP"),
		FLR_DB:            os.Getenv("FLR_DB"),
		FLR_METRICS:       os.Getenv("FLR_METRICS"),
		FLR_OPC:           os.Getenv("FLR_OPC"),
		FLR_FM:            os.Getenv("FLR_FM"),
		PASSWORD:          os.Getenv("PASSWORD"),
		SBS_PASSWORD:      os.Getenv("SBS_PASSWORD"),
		SBS_PASS2:         os.Getenv("SBS_PASS2"),
		SBS_CORE:          os.Getenv("SBS_CORE"),
		SBS_BRIGDE:        os.Getenv("SBS_BRIGDE"),
		SBS_PUPPET:        os.Getenv("SBS_PUPPET"),
		SBS_INTERFACE:     os.Getenv("SBS_INTERFACE"),
		SBS_PLATFORM_API:  os.Getenv("SBS_PLATFORM_API"),
		SBS_PLATFORM_CORE: os.Getenv("SBS_PLATFORM_CORE"),
		SBS_PLATFORM_DB:   os.Getenv("SBS_PLATFORM_DB"),
		SBS_METRICS:       os.Getenv("SBS_METRICS"),
		SBS_STAGING:       os.Getenv("SBS_STAGING"),
		SBS_OPC:           os.Getenv("SBS_OPC"),
	}
}
