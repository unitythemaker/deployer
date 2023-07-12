package web

import (
	"bulut-server/internal/logic/deploy"
	"bulut-server/internal/logic/namespace"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"os"
)

func (s *Server) ConfigureRoutes() {
	s.logger.Info("Configuring routes...")

	deploymentGrp := s.Group("/deployment", s.authMiddleware)
	deploymentGrp.POST("/upload/:namespace/:deployment", s.uploadHandler)

	namespaceGrp := s.Group("/namespace", s.authMiddleware)
	namespaceGrp.POST("/", s.createNamespaceHandler)
}

func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqApiKey := c.Request().Header.Get("authorization")
		if s.config.ApiKey != reqApiKey {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
		}
		return next(c)
	}
}

type CreateNamespaceRequest struct {
	Name string `json:"name" form:"name"`
}

func (s *Server) createNamespaceHandler(c echo.Context) error {
	var req CreateNamespaceRequest
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Bad request",
		})
	}
	if req.Name == "" {
		s.logger.Warn("Missing name in body")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing name in body",
		})
	}

	_, err = namespace.CreateNamespace(s.db, req.Name)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: namespaces.name" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Namespace with this name already exists",
			})
		}
		s.logger.Error(err, "Failed to create namespace")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create namespace",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Namespace created successfully",
	})
}

func (s *Server) uploadHandler(c echo.Context) error {
	entrypoint := c.QueryParam("entrypoint")
	if entrypoint == "" {
		s.logger.Warn("Missing entrypoint query parameter")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing entrypoint query parameter",
		})
	}

	namespaceId := c.Param("namespace")
	if namespaceId == "" {
		s.logger.Warn("Missing namespace-uuid path parameter")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing namespace-uuid path parameter",
		})
	}
	deploymentId := c.Param("deployment")

	wholeForm, err := c.MultipartForm()
	if err != nil {
		s.logger.Warn("Failed to parse form-data", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse form-data",
		})
	}
	s.logger.Info("Received deployment request", "form", wholeForm, "entrypoint", entrypoint)
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
		deploy.BuildAndDeploy(namespaceId, deploymentId, tempFilename, entrypoint, s.logger)
	}()

	response := map[string]string{
		"message": "Build in Progress",
		"file":    tempFilename,
	}

	return c.JSON(http.StatusOK, response)
}
