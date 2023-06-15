package deploy

import (
	"archive/zip"
	"bulut-server/pkg/logger"
	"bytes"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

const DEFAULT_DEPLOY_PORT = 1234

func GenerateTempFilename() string {
	rand.Seed(time.Now().UnixNano())
	randID := fmt.Sprintf("%016x", rand.Uint64())
	return filepath.Join(os.TempDir(), "app-"+randID)
}

func ExtractZip(filePath string, tempDir string) error {
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

func CreateDockerfileIfNotPresent(tempDir string, entrypoint string) error {
	if entrypoint == "" {
		entrypoint = "server/index.mjs" // default for nitro
	}

	dockerfilePath := filepath.Join(tempDir, "Dockerfile")

	dockerfileTmpl := `FROM {{.BaseImage}}
WORKDIR /app
COPY {{.CopySource}} {{.CopyDestination}}
ENV PORT={{.ExposePort}}
EXPOSE {{.ExposePort}}
CMD [{{.Cmd}}]`

	dockerfileTemplate, err := template.New("Dockerfile").Parse(dockerfileTmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = dockerfileTemplate.Execute(&buf, map[string]string{
		// TODO: Make node version configurable
		"BaseImage":       "node:18-alpine",
		"CopySource":      ".",
		"CopyDestination": "./",
		"ExposePort":      "8080",
		"Cmd":             `"node", "` + entrypoint + `"`,
	})
	if err != nil {
		return err
	}

	err = os.WriteFile(dockerfilePath, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func BuildDockerImage(tempDir string) (string, error) {
	imageName := fmt.Sprintf("a-bulut-image:%s", time.Now().Format("20060102150405"))
	buildOpts := docker.BuildImageOptions{
		Name:         imageName,
		ContextDir:   tempDir,
		Dockerfile:   "Dockerfile",
		OutputStream: os.Stdout,
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return "", err
	}

	err = client.BuildImage(buildOpts)
	if err != nil {
		return "", err
	}

	return imageName, nil
}

func DeployDockerContainer(imageName string, containerName string) (string, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return "", err
	}

	containerConfig := docker.Config{
		Image: imageName,
		ExposedPorts: map[docker.Port]struct{}{
			"8080/tcp": {},
		},
	}

	ip, err := findAvailableIP(DEFAULT_DEPLOY_PORT)
	if err != nil {
		return "", err
	}

	hostConfig := docker.HostConfig{
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8080/tcp": {
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

func CleanupResources(tempDir string, filepath string) error {
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

func BuildAndDeploy(filePath string, entrypoint string, logger *logger.Logger) {
	startTime := time.Now().UnixMilli()
	logger.Info("Building and deploying app", "file", filePath)

	containerName := "app-" + fmt.Sprintf("%016x", rand.Uint64())
	tempDir := filepath.Join(os.TempDir(), containerName)
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		logger.Error(err, "Failed to create temporary directory")
		return
	}

	err = ExtractZip(filePath, tempDir)
	if err != nil {
		logger.Error(err, "Failed to extract archive")
		return
	}

	if err := CreateDockerfileIfNotPresent(tempDir, entrypoint); err != nil {
		logger.Error(err, "Failed to create Dockerfile")
		return
	}

	imageName, err := BuildDockerImage(tempDir)
	if err != nil {
		logger.Error(err, "Failed to build Docker image")
		return
	}

	ip, err := DeployDockerContainer(imageName, containerName)
	if err != nil {
		logger.Error(err, "Failed to deploy Docker container")
		return
	}
	logger.Info("Successfully deployed app", "ip", ip, "port", DEFAULT_DEPLOY_PORT, "time_ms", time.Now().UnixMilli()-startTime)

	if err := CleanupResources(tempDir, filePath); err != nil {
		logger.Error(err, "Failed to clean up resources")
	}
}
