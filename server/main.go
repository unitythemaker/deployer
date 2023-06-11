package main

import (
	"Deployer/internal/utils/config"
	"Deployer/internal/utils/logger"
	"Deployer/internal/web"
)

func main() {
	config.Init()

	webServerConfig := config.GetWebServerConfig()

	log := logger.New(logger.Options{
		Development: true,
		Level:       0,
	})

	server := web.NewServer(webServerConfig, log)
	server.Start()
}
