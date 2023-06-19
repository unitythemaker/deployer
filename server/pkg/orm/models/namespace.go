package models

type Namespace struct {
	BaseModel
	Name        string       `gorm:"unique;not null"`
	Deployments []Deployment `gorm:"foreignKey:NamespaceID"`
}
