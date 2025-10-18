package models

import "gorm.io/gorm"

type Subject struct {
	gorm.Model
	Name  string `json:"subject_name"`
	C_hrs int16  `json:"credit_hours"`
}

func CreateSubject(db *gorm.DB, chrs int, name string) (*Subject, error) {
	subject := Subject{
		Name:  name,
		C_hrs: int16(chrs),
	}

	if err := db.Create(&subject).Error; err != nil {
		return nil, err
	}
	return &subject, nil
}
