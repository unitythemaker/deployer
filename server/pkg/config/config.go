package config

import (
	"bulut-server/internal/web"
	"bulut-server/pkg/orm/common"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
)

func Init() {
	err := InitViper()
	if err != nil {
		log.Fatalf("Failed to initialize viper: %s", err)
	}
	err = InitDotenv()
	if err != nil {
		log.Printf("Failed to load .env file: %s", err)
	}
}

func InitViper() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	// Set default values
	viper.SetDefault("myKey", "defaultValue")
	if err := viper.ReadInConfig(); err != nil {
		// If the config file is not found, do not return an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	return nil
}

func InitDotenv() error {
	err := godotenv.Load()
	if err != nil {
		// if .env file is not found, ignore the error
		if _, ok := err.(*os.PathError); !ok {
			return err
		}
	}
	return nil
}

func GetWebServerConfig() *web.ServerConfig {
	host := os.Getenv("HOST")
	portStr := os.Getenv("PORT")
	apiKey := os.Getenv("API_KEY")

	defaultPort := 8080

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		log.Printf("Invalid or missing port value: %s. Using default port: %d\n", portStr, defaultPort)
		port = defaultPort
	}

	if apiKey == "" {
		log.Fatal("Missing API_KEY environment variable")
	}

	return &web.ServerConfig{
		Host:   host,
		Port:   port,
		ApiKey: apiKey,
	}
}

func GetDatabaseConfig() *common.DatabaseConfig {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "bulut.db"
	}
	return &common.DatabaseConfig{
		DBPath: dbPath,
	}
}
