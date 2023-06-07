package config

import (
	"NitroDeployer/internal/web"
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
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func InitDotenv() {
	err := godotenv.Load()
	if err != nil {
		// if .env file is not found, ignore the error
		if _, ok := err.(*os.PathError); !ok {
			log.Fatal(err)
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
