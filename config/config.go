package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SSHUser     string
	SSHPassword string
	ServerFLR   string
	ServerSBS   string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return Config{
		SSHUser:     os.Getenv("SSH_USER"),
		SSHPassword: os.Getenv("SSH_PASSWORD"),
		ServerFLR:   os.Getenv("SERVER_FLR"),
		ServerSBS:   os.Getenv("SERVER_SBS"),
	}
}
