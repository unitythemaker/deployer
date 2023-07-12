package web

import (
	"bulut-server/pkg/logger"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ServerConfig struct {
	Host   string
	Port   int
	ApiKey string
}

type Server struct {
	config       *ServerConfig
	logger       *logger.Logger
	db           *gorm.DB
	dockerClient *docker.Client
	*echo.Echo
}

type ServerUtils struct {
	Logger       *logger.Logger
	DockerClient *docker.Client
	Db           *gorm.DB
}

func NewServer(config *ServerConfig, components ServerUtils) *Server {
	s := &Server{
		config:       config,
		logger:       components.Logger,
		db:           components.Db,
		dockerClient: components.DockerClient,
		Echo:         echo.New(),
	}
	s.ConfigureRoutes()
	// Disabled due to last params having a bug, an unwanted slash is added
	//s.Echo.Pre(middleware.AddTrailingSlash())
	return s
}
