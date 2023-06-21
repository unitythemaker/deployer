package web

import (
	"bulut-server/pkg/logger"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/labstack/echo/v4"
)

type ServerConfig struct {
	Host   string
	Port   int
	ApiKey string
}

type Server struct {
	config       *ServerConfig
	logger       *logger.Logger
	dockerClient *docker.Client
	*echo.Echo
}

func NewServer(config *ServerConfig, logger *logger.Logger) *Server {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		logger.Error(err, "Failed to create docker client")
	}

	s := &Server{
		config:       config,
		logger:       logger,
		dockerClient: dockerClient,
		Echo:         echo.New(),
	}
	s.ConfigureRoutes()
	return s
}
