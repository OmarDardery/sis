package database

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDb() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(os.Getenv("DB_DSN")))
}
