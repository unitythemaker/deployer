package deploy

import (
	"bulut-server/pkg/orm/models"
	"gorm.io/gorm"
)

func FindDeploymentByName(db *gorm.DB, name, namespaceId string) (models.Deployment, error) {
	var deployment models.Deployment
	result := db.Where("name = ? AND namespace_id = ?", name, namespaceId).First(&deployment)
	return deployment, result.Error
}
