package web

import (
	"Deployer/internal/utils/logger"
	"archive/zip"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const DEFAULT_DEPLOY_PORT = 1234

type ServerConfig struct {
	Host string
	Port int
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
		log.Fatalf("Failed to create docker client: %s", err)
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

func (s *Server) ConfigureRoutes() {
	s.logger.Info("Configuring routes...")

	s.POST("/upload", s.uploadHandler)

}

func (s *Server) Start() {
	s.logger.Info("Connecting to Docker...")
	dockerInfo, err := s.dockerClient.Info()
	if err != nil {
		log.Fatalf("Failed to get docker info: %s", err)
	}
	s.logger.Info("Connected to Docker", "version", dockerInfo.ServerVersion)
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.logger.Info("Starting server on", "address", address)
	err = s.Echo.Start(address)
	if err != nil {
		log.Fatal(err)
	}
}

func generateTempFilename() string {
	rand.Seed(time.Now().UnixNano())
	randID := fmt.Sprintf("%016x", rand.Uint64())
	return filepath.Join(os.TempDir(), "app-"+randID)
}

func (s *Server) uploadHandler(c echo.Context) error {
	wholeForm, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse form-data",
		})
	}
	s.logger.Info("Received form-data", "form", wholeForm)
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
	// Print the size of the received file

	tempFilename := generateTempFilename()
	dst, err := os.Create(tempFilename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create temporary file",
		})
	}
	defer dst.Close()
	// Print the size of the received file
	fileSize, err := dst.Stat()
	if err == nil {
		fmt.Printf("Size of the file received: %d\n", fileSize.Size())
	}

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
	startTime := time.Now().UnixMilli()
	s.logger.Info("Building and deploying app using", "file", filePath)

	containerName := "app-" + fmt.Sprintf("%016x", rand.Uint64())
	tempDir := filepath.Join(os.TempDir(), containerName)
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		s.logger.Error(err, "Failed to create temporary directory")
		return
	}

	err = s.extractZip(filePath, tempDir)
	if err != nil {
		s.logger.Error(err, "Failed to extract archive")
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

	ip, err := s.deployDockerContainer(imageName, containerName)
	if err != nil {
		s.logger.Error(err, "Failed to deploy Docker container")
		return
	}
	s.logger.Info("Successfully deployed app", "ip", ip, "port", DEFAULT_DEPLOY_PORT, "time_ms", time.Now().UnixMilli()-startTime)

	if err := s.cleanupResources(tempDir, filePath); err != nil {
		s.logger.Error(err, "Failed to clean up resources")
	}
}

func (s *Server) extractZip(filePath string, tempDir string) error {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		destPath := filepath.Join(tempDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
		} else {
			srcFile, err := file.Open()
			if err != nil {
				return err
			}
			defer srcFile.Close()

			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) createDockerfileIfNotPresent(tempDir string) error {
	dockerfilePath := filepath.Join(tempDir, "Dockerfile")

	// TODO: possible security issue here, if the user uploads a Dockerfile, it must be overwritten
	//if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
	dockerfileContent := []byte("FROM node:18-alpine\n" +
		"WORKDIR /app\n" +
		"COPY .output/ ./\n" +
		"EXPOSE 8080\n" +
		"CMD [\"node\", \"./server/index.mjs\"]\n")
	err := os.WriteFile(dockerfilePath, dockerfileContent, 0644)
	if err != nil {
		return err
	}
	//}

	return nil
}

func (s *Server) buildDockerImage(tempDir string) (string, error) {
	imageName := fmt.Sprintf("a-deployer-image:%s", time.Now().Format("20060102150405"))
	buildOpts := docker.BuildImageOptions{
		Name:         imageName,
		ContextDir:   tempDir,
		Dockerfile:   "Dockerfile",
		OutputStream: os.Stdout,
	}

	err := s.dockerClient.BuildImage(buildOpts)
	if err != nil {
		return "", err
	}

	return imageName, nil
}

func (s *Server) deployDockerContainer(imageName string, containerName string) (string, error) {
	//manifest := []byte("YOUR_MANIFEST_CONTENT_HERE")
	//manifestFilePath := filepath.Join("path/to/your/manifest/folder", "app-deployment.yml")
	//
	//err := os.WriteFile(manifestFilePath, manifest, 0644)
	//if err != nil {
	//	return err
	//}
	//
	//cmd := exec.Command("kubectl", "apply", "-f", manifestFilePath)
	//err = cmd.Run()
	//if err != nil {
	//	return err
	//}
	// instead of Kubernetes, we will use Docker
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return "", err
	}

	containerConfig := docker.Config{
		Image: imageName,
		ExposedPorts: map[docker.Port]struct{}{
			"3000/tcp": {},
		},
	}

	ip, err := findAvailableIP(DEFAULT_DEPLOY_PORT)
	if err != nil {
		return "", err
	}

	hostConfig := docker.HostConfig{
		PortBindings: map[docker.Port][]docker.PortBinding{
			"3000/tcp": {
				{
					HostIP:   ip,
					HostPort: strconv.Itoa(DEFAULT_DEPLOY_PORT),
				},
			},
		},
	}

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name:       containerName,
		Config:     &containerConfig,
		HostConfig: &hostConfig,
	})
	if err != nil {
		return "", err
	}

	err = client.StartContainer(container.ID, nil)
	if err != nil {
		return "", err
	}

	return ip, nil
}

func (s *Server) cleanupResources(tempDir string, filepath string) error {
	err := os.RemoveAll(tempDir)
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath)
	if err != nil {
		return err
	}

	return nil
}

func findAvailableIP(port int) (string, error) {
	for i := 100; i <= 255; i++ {
		ip := fmt.Sprintf("127.0.0.%d", i)
		addr := fmt.Sprintf("%s:%d", ip, port)

		listener, err := net.Listen("tcp", addr)

		if err == nil {
			_ = listener.Close()
			return ip, nil
		}
	}

	return "", fmt.Errorf("no available IP found")
}
