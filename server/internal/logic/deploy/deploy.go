package deploy

import (
	"archive/zip"
	"bulut-server/internal/logic/revision"
	"bulut-server/pkg/logger"
	"bulut-server/pkg/orm/models"
	"bytes"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

func CreateDockerfileIfNotPresent(tempDir, entrypoint string) error {
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

type ImageBuildResult struct {
	ImageID   string
	ImageName string
	ImageTag  string
	Image     *docker.Image
}

func BuildDockerImage(imageRepo, tempDir string) (*ImageBuildResult, error) {
	imageTag := time.Now().Format("20060102150405")
	imageName := fmt.Sprintf("%s:%s", imageRepo, imageTag)
	buildOpts := docker.BuildImageOptions{
		Name:         imageName,
		ContextDir:   tempDir,
		Dockerfile:   "Dockerfile",
		OutputStream: os.Stdout,
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	err = client.BuildImage(buildOpts)
	if err != nil {
		return nil, err
	}

	// Tag image as latest
	err = client.TagImage(imageName, docker.TagImageOptions{
		Repo:  imageRepo,
		Tag:   "latest",
		Force: true,
	})
	if err != nil {
		return nil, err
	}

	image, err := client.InspectImage(imageName)
	if err != nil {
		return nil, err
	}

	return &ImageBuildResult{
		ImageID:   image.ID,
		ImageName: imageName,
		ImageTag:  imageTag,
		Image:     image,
	}, nil
}

type ContainerDeployResult struct {
	ContainerID string
	Container   *docker.Container
	IP          string // Deprecated: Gateway/Domains will be used instead
}

func DeployDockerContainer(imageName, containerName string) (*ContainerDeployResult, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	containerConfig := docker.Config{
		Image: imageName,
		ExposedPorts: map[docker.Port]struct{}{
			"8080/tcp": {},
		},
	}

	ip, err := findAvailableIP(DEFAULT_DEPLOY_PORT)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	err = client.StartContainer(container.ID, nil)
	if err != nil {
		return nil, err
	}

	return &ContainerDeployResult{
		ContainerID: container.ID,
		Container:   container,
		IP:          ip,
	}, nil
}

func CleanupResources(tempDir, filepath string) error {
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

// Deprecated: Gateway/Domains will be used instead
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

type BuildAndDeployOpts struct {
	NamespaceId  string
	DeploymentId uuid.UUID
	FilePath     string
	Entrypoint   string
	Logger       *logger.Logger
	Db           *gorm.DB
}

func BuildAndDeploy(opts BuildAndDeployOpts) {
	logger := opts.Logger
	db := opts.Db
	startTime := time.Now().UnixMilli()
	logger.Info("Building and deploying app", "file", opts.FilePath)

	dockerName := fmt.Sprintf("bulut-%s-%s", opts.NamespaceId, opts.DeploymentId)
	tempDir := filepath.Join(os.TempDir(), dockerName)
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		logger.Error(err, "Failed to create temporary directory")
		return
	}

	defer func(tempDir, filepath string) {
		err := CleanupResources(tempDir, filepath)
		if err != nil {
			logger.Error(err, "Failed to cleanup resources")
			return
		}
	}(tempDir, opts.FilePath)

	err = ExtractZip(opts.FilePath, tempDir)
	if err != nil {
		logger.Error(err, "Failed to extract archive")
		return
	}

	if err := CreateDockerfileIfNotPresent(tempDir, opts.Entrypoint); err != nil {
		logger.Error(err, "Failed to create Dockerfile")
		return
	}

	buildResult, err := BuildDockerImage(dockerName, tempDir)
	if err != nil {
		logger.Error(err, "Failed to build Docker image")
		return
	}
	var currentDeployment models.Deployment
	err = db.Where("id = ?", opts.DeploymentId).First(&currentDeployment).Error
	if err != nil {
		logger.Error(err, "Failed to get current deployment")
		return
	}
	_, err = revision.CreateRevision(db, opts.DeploymentId, dockerName, buildResult.ImageTag, buildResult.ImageID)

	// Delete old container
	if currentDeployment.ContainerID != "" {
		if err := DeleteContainer(currentDeployment.ContainerID); err != nil {
			logger.Error(err, "Failed to delete old container")
			return
		}
	}

	imageName := fmt.Sprintf("%s:%s", dockerName, buildResult.ImageTag)
	deployResult, err := DeployDockerContainer(imageName, dockerName)
	if err != nil {
		logger.Error(err, "Failed to deploy Docker container")
		return
	}
	logger.Info("Successfully deployed app", "ip", deployResult.IP, "port", DEFAULT_DEPLOY_PORT, "time_ms", time.Now().UnixMilli()-startTime)

	// Update deployment
	currentDeployment.ContainerID = deployResult.ContainerID
	if err := db.Save(&currentDeployment).Error; err != nil {
		logger.Error(err, "Failed to update deployment")
		return
	}
}

func DeleteContainer(containerID string) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	err = client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            containerID,
		RemoveVolumes: true, // TOOD: Possible data loss
		Force:         true,
	})
	if err != nil {
		return err
	}

	return nil
}
