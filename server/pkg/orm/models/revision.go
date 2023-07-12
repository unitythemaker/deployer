package models

import "github.com/google/uuid"

type Revision struct {
	BaseModel
	DeploymentID uuid.UUID `gorm:"not null"`
	Deployment   Deployment
	ImageName    string `gorm:"not null"`
	ImageTag     string `gorm:"not null"`
	ImageID      string `gorm:"not null"`
	Notes        string
}
