package namespace

import (
	"bulut-server/pkg/orm/models"
	"gorm.io/gorm"
)

func CreateNamespace(db *gorm.DB, name string) (models.Namespace, error) {
	namespace := models.Namespace{Name: name}
	result := db.Create(&namespace)
	return namespace, result.Error
}
