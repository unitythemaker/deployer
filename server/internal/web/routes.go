package web

import (
	"bulut-server/internal/logic/deploy"
	"bulut-server/internal/logic/namespace"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
)

func (s *Server) ConfigureRoutes() {
	s.logger.Info("Configuring routes...")

	deploymentGrp := s.Group("/deployment", s.authMiddleware)
	deploymentGrp.GET("/:namespace/:deployment", s.getDeploymentHandler)
	deploymentGrp.POST("/", s.createDeploymentHandler)
	deploymentGrp.PUT("/upload/:namespace/:deployment", s.uploadHandler)

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

func (s *Server) getDeploymentHandler(c echo.Context) error {
	namespaceName := c.Param("namespace")
	deploymentName := c.Param("deployment")

	dep, err := deploy.FindDeploymentByName(s.db, namespaceName, deploymentName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Deployment not found",
			})
		}
		s.logger.Error(err, "Failed to get deployment")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get deployment",
		})
	}

	return c.JSON(http.StatusOK, dep)
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

type CreateDeploymentRequest struct {
	Name      string `json:"name" form:"name"`
	Namespace string `json:"namespace" form:"namespace"`
}

func (s *Server) createDeploymentHandler(c echo.Context) error {
	var req CreateDeploymentRequest
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
	} else if req.Namespace == "" {
		s.logger.Warn("Missing namespace in body")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing namespace in body",
		})
	} else if len(req.Name) < 4 {
		s.logger.Warn("Name is too short")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name is too short",
		})
	} else if len(req.Name) > 50 {
		s.logger.Warn("Name is too long")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name is too long",
		})
	}

	namespace, err := namespace.FindNamespaceByName(s.db, req.Namespace)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Namespace not found",
			})
		}
		s.logger.Error(err, "Failed to find namespace")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to find namespace",
		})
	}

	_, err = deploy.CreateDeployment(s.db, req.Name, namespace.ID)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: deployments.name, deployments.namespace_id" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Deployment with this name already exists",
			})
		}
		s.logger.Error(err, "Failed to create deployment")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create deployment",
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Deployment created successfully",
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

	namespace, err := namespace.FindNamespaceByName(s.db, c.Param("namespace"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Namespace not found",
			})
		}
		s.logger.Error(err, "Failed to find namespace")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to find namespace",
		})
	}
	namespaceId := namespace.ID.String()

	deployment, err := deploy.FindDeploymentByName(s.db, c.Param("deployment"), namespaceId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Deployment not found",
			})
		}
		s.logger.Error(err, "Failed to find deployment")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to find deployment",
		})
	}
	deploymentId := deployment.ID

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
		deploy.BuildAndDeploy(deploy.BuildAndDeployOpts{
			NamespaceId:  namespaceId,
			DeploymentId: deploymentId,
			FilePath:     tempFilename,
			Entrypoint:   entrypoint,
			Db:           s.db,
			Logger:       s.logger,
		})
	}()

	response := map[string]string{
		"message": "Build in Progress",
		"file":    tempFilename,
	}

	return c.JSON(http.StatusOK, response)
}
