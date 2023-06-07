package main

import (
	"NitroDeployer/internal/utils/config"
	"NitroDeployer/internal/utils/logger"
	"NitroDeployer/internal/web"
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
