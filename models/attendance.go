package models

import "gorm.io/gorm"

type Attendance struct {
	gorm.Model
	SlotID    uint    `json:"slot_id"` // Foreign key to Slot
	Slot      Slot    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"slot"`
	StudentID uint    `json:"student_id"` // Foreign key to Student
	Student   Student `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"student"`
	Status    string  `json:"status"` // e.g. "present", "absent", "late"
}
