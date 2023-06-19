package common

import (
	"bulut-server/pkg/orm/models"
	"fmt"
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

func (q *DeploymentQuery) GetDeploymentByContainerID(namespace uuid.UUID, id string) (*models.Deployment, error) {
	var deployment models.Deployment
	result := q.gdb.Where("namespace_id = ? AND container_id = ?", namespace, id).Limit(1).First(&deployment)
	fmt.Println(3, result)
	if result.Error != nil {
		return nil, result.Error
	}
	return &deployment, nil
}

func (q *DeploymentQuery) CreateDeployment(deployment *models.Deployment) error {
	return q.gdb.Create(deployment).Error
}
