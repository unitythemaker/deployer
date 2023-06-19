package common

import (
	"bulut-server/pkg/orm/models"
	"gorm.io/gorm"
)

type NamespaceQuery struct {
	db  *DatabaseController
	gdb *gorm.DB
}

func NewNamespaceQuery(db *DatabaseController) *NamespaceQuery {
	return &NamespaceQuery{
		db:  db,
		gdb: db.gdb,
	}
}

func (q *NamespaceQuery) GetNamespaceByName(name string) (*models.Namespace, error) {
	var namespace models.Namespace
	err := q.gdb.Where("name = ?", name).Limit(1).First(&namespace).Error
	if err != nil {
		return nil, err
	}
	return &namespace, nil
}

func (q *NamespaceQuery) CreateNamespace(namespace *models.Namespace) error {
	return q.gdb.Create(namespace).Error
}
