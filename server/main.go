package main

import (
	"Deployer/internal/utils/config"
	"Deployer/internal/utils/logger"
	"Deployer/internal/web"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.Init()

	webServerConfig := config.GetWebServerConfig()

	log := logger.New(logger.Options{
		Development: true,
		Level:       logger.InfoLevel,
	})

	server := web.NewServer(webServerConfig, log)

	// Add middleware for gracefully handling panics
	server.Use(middleware.Recover())

	server.Start()
}
