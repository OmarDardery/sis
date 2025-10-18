package models

import (
	"github.com/OmarDardery/sis/database"
	"gorm.io/gorm"
)

func Migrate() *gorm.DB {
	db, err := database.NewDb()
	if err != nil {
		panic(err.Error())
	}

	err = db.AutoMigrate(&Student{}, &Teacher{}, &Subject{}, &Semester{}, &Slot{}, &Attendance{})
	if err != nil {
		panic(err.Error())
	}

	return db
}
