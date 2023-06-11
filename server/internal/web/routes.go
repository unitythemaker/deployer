package web

import (
	"Deployer/internal/deploy"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"os"
)

func (s *Server) ConfigureRoutes() {
	s.logger.Info("Configuring routes...")

	s.POST("/upload", s.uploadHandler)

}

func (s *Server) uploadHandler(c echo.Context) error {
	wholeForm, err := c.MultipartForm()
	if err != nil {
		s.logger.Warn("Failed to parse form-data", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse form-data",
		})
	}
	s.logger.Info("Received form-data", "form", wholeForm)
	file, err := c.FormFile("file")
	if err != nil {
		s.logger.Warn("Failed to retrieve file from form-data", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to retrieve file from form-data",
		})
	}

	src, err := file.Open()
	if err != nil {
		s.logger.Warn("Failed to open file", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	tempFilename := deploy.GenerateTempFilename()
	dst, err := os.Create(tempFilename)
	if err != nil {
		s.logger.Warn("Failed to create temporary file", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create temporary file",
		})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		s.logger.Error(err, "Failed to save uploaded file")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save uploaded file",
		})
	}

	go func() {
		deploy.BuildAndDeploy(tempFilename, s.dockerClient, s.logger)
	}()

	response := map[string]string{
		"message": "Build in Progress",
		"file":    tempFilename,
	}

	return c.JSON(http.StatusOK, response)
}
