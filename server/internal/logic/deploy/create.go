package deploy

import (
	"bulut-server/pkg/orm/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateDeployment(db *gorm.DB, name string, namespaceId uuid.UUID) (models.Deployment, error) {
	deployment := models.Deployment{
		Name:        name,
		NamespaceID: namespaceId,
	}
	result := db.Create(&deployment)
	return deployment, result.Error
}
