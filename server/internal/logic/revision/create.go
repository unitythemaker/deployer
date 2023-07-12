package revision

import (
	"bulut-server/pkg/orm/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateRevision(db *gorm.DB, deploymentId uuid.UUID, imageName, imageTag, imageID string) (models.Revision, error) {
	revision := models.Revision{
		DeploymentID: deploymentId,
		ImageName:    imageName,
		ImageTag:     imageTag,
		ImageID:      imageID,
	}
	result := db.Create(&revision)
	return revision, result.Error
}
