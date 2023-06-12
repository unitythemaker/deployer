package main

import (
	"Deployer/internal/web"
	"Deployer/pkg/config"
	"Deployer/pkg/logger"
	"github.com/labstack/echo/v4/middleware"
	"strconv"
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

	serverConfig := config.GetWebServerConfig()
	portStr := strconv.Itoa(serverConfig.Port)
	address := serverConfig.Host + ":" + portStr
	server.Start(address)
}
