package models

import "gorm.io/gorm"

type Semester struct {
	gorm.Model
	Title string `json:"semester_title"`
}

func CreateSemester(db *gorm.DB, title string) (*Semester, error) {
	semester := Semester{
		Title: title,
	}

	if err := db.Create(&semester).Error; err != nil {
		return nil, err
	}
	return &semester, nil
}
