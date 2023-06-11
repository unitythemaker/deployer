package config

import (
	"Deployer/internal/web"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
)

func Init() {
	InitViper()
	InitDotenv()
}

func InitViper() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	// Set default values
	viper.SetDefault("myKey", "defaultValue")
	if err := viper.ReadInConfig(); err != nil {
		// If the config file is not found, do not return an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Failed to read config file: %s", err)
		}
	}
}

func InitDotenv() {
	err := godotenv.Load()
	if err != nil {
		// if .env file is not found, ignore the error
		if _, ok := err.(*os.PathError); !ok {
			log.Fatalf("Failed to load .env file: %s", err)
		}
	}
}

func GetWebServerConfig() *web.ServerConfig {
	host := os.Getenv("HOST")
	portStr := os.Getenv("PORT")

	defaultPort := 8080

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		log.Printf("Invalid or missing port value: %s. Using default port: %d\n", portStr, defaultPort)
		port = defaultPort
	}

	return &web.ServerConfig{
		Host: host,
		Port: port,
	}
}
