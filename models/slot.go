package models

import (
	"time"

	"gorm.io/gorm"
)

type Slot struct {
	gorm.Model
	TeacherID  uint
	SubjectID  uint
	SemesterID uint

	Date      time.Time `gorm:"type:date;not null"` // just date (YYYY-MM-DD)
	StartTime time.Time `gorm:"type:time;not null"` // time only (HH:MM:SS)
	EndTime   time.Time `gorm:"type:time;not null"`

	// Relationships
	Teacher  Teacher  `gorm:"foreignKey:TeacherID"`
	Subject  Subject  `gorm:"foreignKey:SubjectID"`
	Semester Semester `gorm:"foreignKey:SemesterID"`
}

func CreateSlot(db *gorm.DB, teacher *Teacher, subject *Subject, semester *Semester, date time.Time, start time.Time, end time.Time) (*Slot, error) {
	slot := Slot{
		TeacherID:  teacher.ID,
		SubjectID:  subject.ID,
		SemesterID: semester.ID,
		Date:       date,
		StartTime:  start,
		EndTime:    end,
	}

	if err := db.Create(&slot).Error; err != nil {
		return nil, err
	}
	return &slot, nil
}
