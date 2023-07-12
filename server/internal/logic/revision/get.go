package revision

import (
	"bulut-server/pkg/orm/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetLatestRevision(db *gorm.DB, deploymentId uuid.UUID) (models.Revision, error) {
	var revision models.Revision
	result := db.Where("deployment_id = ?", deploymentId).Order("created_at desc").First(&revision)
	return revision, result.Error
}
