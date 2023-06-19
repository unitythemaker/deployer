package common

import (
	"bulut-server/pkg/orm/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	DBPath string
}

type DatabaseController struct {
	Config *DatabaseConfig
	gdb    *gorm.DB
}

func NewDatabase(config *DatabaseConfig) *DatabaseController {
	return &DatabaseController{
		Config: config,
	}
}

func (db *DatabaseController) Connect() error {
	gdb, err := gorm.Open(sqlite.Open(db.Config.DBPath), &gorm.Config{})
	if err != nil {
		return err
	}

	db.gdb = gdb

	return nil
}

func (db *DatabaseController) AutoMigrate() error {
	return db.gdb.AutoMigrate(&models.Namespace{}, &models.Deployment{})
}
