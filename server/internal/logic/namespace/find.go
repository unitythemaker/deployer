package namespace

import (
	"bulut-server/pkg/orm/models"
	"gorm.io/gorm"
)

func FindNamespaceByName(db *gorm.DB, name string) (models.Namespace, error) {
	var namespace models.Namespace
	result := db.Where("name = ?", name).First(&namespace)
	return namespace, result.Error
}
