package main

import (
	"bulut-server/internal/web"
	"bulut-server/pkg/config"
	"bulut-server/pkg/logger"
	"bulut-server/pkg/orm/common"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/labstack/echo/v4/middleware"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	config.Init()

	webServerConfig := config.GetWebServerConfig()

	log := logger.New(logger.Options{
		Development: true,
		Level:       logger.InfoLevel,
	})

	dbConfig := config.GetDatabaseConfig()
	db, err := common.ConnectDB(dbConfig)
	if err != nil {
		log.Error(err, "Failed to connect to database")
	}
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		log.Error(err, "Failed to connect to docker")
	}

	server := web.NewServer(webServerConfig, web.ServerUtils{
		Logger:       log,
		DockerClient: dockerClient,
		Db:           db,
	})

	// Add middleware for gracefully handling panics
	server.Use(middleware.Recover())

	serverConfig := config.GetWebServerConfig()
	portStr := strconv.Itoa(serverConfig.Port)
	address := serverConfig.Host + ":" + portStr
	err = server.Start(address)
	if err != nil {
		log.Error(err, "Failed to start server")
	}

	// graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	<-stopChan
}
