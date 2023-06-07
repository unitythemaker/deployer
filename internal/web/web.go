package web

import (
	"NitroDeployer/internal/utils/logger"
	"archive/zip"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ServerConfig struct {
	Host string
	Port int
}

type Server struct {
	config *ServerConfig
	logger *logger.Logger
	*echo.Echo
}

func NewServer(config *ServerConfig, logger *logger.Logger) *Server {
	s := &Server{
		config: config,
		logger: logger,
		Echo:   echo.New(),
	}
	s.ConfigureRoutes()
	return s
}

func (s *Server) ConfigureRoutes() {
	s.logger.Info("Configuring routes...")

	// Upload endpoint
	s.POST("/upload", s.uploadHandler)

}

func (s *Server) Start() {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Info("Starting server on", "address", address)
	err := s.Echo.Start(address)
	if err != nil {
		log.Fatal(err)
	}
}

func generateTempFilename() string {
	rand.Seed(time.Now().UnixNano())
	randID := fmt.Sprintf("%016x", rand.Uint64())
	return filepath.Join(os.TempDir(), "app-"+randID+".zip")
}

func (s *Server) uploadHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to retrieve file from form-data",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	tempFilename := generateTempFilename()
	dst, err := os.Create(tempFilename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create temporary file",
		})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save uploaded file",
		})
	}

	go s.buildAndDeploy(tempFilename)

	response := map[string]string{
		"message": "Build in Progress",
		"file":    tempFilename,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) buildAndDeploy(filePath string) {
	s.logger.Info("Building and deploying app using", "file", filePath)

	tempDir, err := s.extractZip(filePath)
	if err != nil {
		s.logger.Error(err, "Failed to extract zip")
		return
	}

	if err := s.createDockerfileIfNotPresent(tempDir); err != nil {
		s.logger.Error(err, "Failed to create Dockerfile")
		return
	}

	imageName, err := s.buildDockerImage(tempDir)
	if err != nil {
		s.logger.Error(err, "Failed to build Docker image")
		return
	}

	if err := s.deployDockerContainer(imageName); err != nil {
		s.logger.Error(err, "Failed to deploy Docker container")
		return
	}

	if err := s.cleanupResources(tempDir); err != nil {
		s.logger.Error(err, "Failed to clean up resources")
	}
}

func (s *Server) extractZip(filePath string) (string, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	tempDir := os.TempDir()

	for _, file := range reader.File {
		destPath := filepath.Join(tempDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
		} else {
			srcFile, err := file.Open()
			if err != nil {
				return "", err
			}
			defer srcFile.Close()

			destFile, err := os.Create(destPath)
			if err != nil {
				return "", err
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return "", err
			}
		}
	}

	return tempDir, nil
}

func (s *Server) createDockerfileIfNotPresent(tempDir string) error {
	// Implement the function to create the Dockerfile if it doesn't exist in the extracted folder.
}

func (s *Server) buildDockerImage(tempDir string) (string, error) {
	// Implement the function to build the Docker image using "go-dockerclient" library.
}

func (s *Server) deployDockerContainer(imageName string) error {
	// Implement the function to deploy the Docker container using the Kubernetes API or `kubectl`.
}

func (s *Server) cleanupResources(tempDir string) error {
	// Implement the function to clean up the temporary files and resources.
}
