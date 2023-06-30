package common

import (
	"bulut-server/pkg/orm/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	DBPath string
}

func ConnectDB(config *DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Namespace{}, &models.Deployment{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
