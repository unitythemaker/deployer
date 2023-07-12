package models

import "github.com/google/uuid"

type Deployment struct {
	BaseModel
	Name        string `gorm:"unique;not null"`
	ContainerID string
	NamespaceID uuid.UUID `gorm:"not null"`
	Namespace   Namespace `gorm:"foreignKey:NamespaceID"`
	Revisions   []Revision
}
