package models

import "gorm.io/gorm"

type Student struct {
	gorm.Model
	Password string `json:"student_password"`
	Email    string `json:"student_email"`
}

func CreateStudent(db *gorm.DB, password, email string) (*Student, error) {
	student := Student{
		Password: password,
		Email:    email,
	}

	if err := db.Create(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
