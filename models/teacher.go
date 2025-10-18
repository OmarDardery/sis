package models

import "gorm.io/gorm"

type Teacher struct {
	gorm.Model
	Password string `json:"teacher_password"`
	Email    string `json:"teacher_email"`
}

func CreateTeacher(db *gorm.DB, password, email string) (*Teacher, error) {
	teacher := Teacher{
		Password: password,
		Email:    email,
	}

	if err := db.Create(&teacher).Error; err != nil {
		return nil, err
	}
	return &teacher, nil
}
