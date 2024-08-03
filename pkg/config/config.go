package config

import (
	"os"
	"welcome-to-the-jungle/pkg/models"

	"github.com/anacrolix/log"
	"github.com/joho/godotenv"
)

func LoadConfig() (models.Config, error) {

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	var dataLoaded bool
	if os.Getenv("DATA_LOADED") == "true" {
		dataLoaded = true
	}

	cfg := models.Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		DataLoaded:  dataLoaded,
		CSVFilePath: os.Getenv("CSV_FILE_PATH"),
	}

	return cfg, nil
}
