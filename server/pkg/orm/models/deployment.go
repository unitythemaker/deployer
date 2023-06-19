package models

import "github.com/google/uuid"

type Deployment struct {
	BaseModel
	Name        string    `gorm:"unique;not null"`
	ImageID     string    `gorm:"not null"`
	ImageSize   uint      `gorm:"not null"`
	ContainerID string    `gorm:"not null"`
	NamespaceID uuid.UUID `gorm:"not null"`
	Namespace   Namespace `gorm:"foreignKey:NamespaceID"`
}
