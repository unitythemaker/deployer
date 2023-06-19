package common

import (
	"bulut-server/pkg/orm/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeploymentQuery struct {
	db  *DatabaseController
	gdb *gorm.DB
}

func NewDeploymentQuery(db *DatabaseController) *DeploymentQuery {
	return &DeploymentQuery{
		db:  db,
		gdb: db.gdb,
	}
}

func (q *DeploymentQuery) GetDeployment(namespaceId uuid.UUID, containerId string) (*models.Deployment, error) {
	var deployment models.Deployment
	err := q.gdb.Where("namespace_id = ? AND container_id = ?", namespaceId, containerId).Limit(1).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (q *DeploymentQuery) CreateDeployment(deployment *models.Deployment) error {
	return q.gdb.Create(deployment).Error
}
