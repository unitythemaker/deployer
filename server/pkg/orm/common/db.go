package common

import (
	"bulut-server/pkg/orm/models"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
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
	// TODO: Timezone
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", db.Config.Host, db.Config.User, db.Config.Password, db.Config.Database, db.Config.Port)
	gdb, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
		// TODO: Change to false
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db.gdb = gdb

	return nil
}

func (db *DatabaseController) AutoMigrate() error {
	return db.gdb.AutoMigrate(&models.Namespace{}, &models.Deployment{})
}
