package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/OmarDardery/sis/models"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

func PostSlots(ctx *gin.Context, db *gorm.DB) {
	var input struct {
		TeacherID  uint   `json:"teacher_id"`
		SubjectID  uint   `json:"subject_id"`
		SemesterID uint   `json:"semester_id"`
		Date       string `json:"date"`       // "2025-10-25"
		StartTime  string `json:"start_time"` // "09:00:00"
		EndTime    string `json:"end_time"`   // "11:00:00"
	}
	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var teacher models.Teacher
	var subject models.Subject
	var semester models.Semester

	if err := db.First(&teacher, input.TeacherID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Teacher not found"})
		return
	}
	if err := db.First(&subject, input.SubjectID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Subject not found"})
		return
	}
	if err := db.First(&semester, input.SemesterID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Semester not found"})
		return
	}

	date, _ := time.Parse("2006-01-02", input.Date)
	start, _ := time.Parse("15:04:05", input.StartTime)
	end, _ := time.Parse("15:04:05", input.EndTime)

	slot, err := models.CreateSlot(db, &teacher, &subject, &semester, date, start, end)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, slot)
}

func PostSubject(ctx *gin.Context, db *gorm.DB) {
	var input struct {
		Name string `json:"name"`
		CHrs int    `json:"credit_hours"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	subject, err := models.CreateSubject(db, input.CHrs, input.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, subject)
}

func PostSemester(ctx *gin.Context, db *gorm.DB) {
	var input struct {
		Title string `json:"title"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	semester, err := models.CreateSemester(db, input.Title)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, semester)
}

// GenerateQRCode returns a QR PNG that encodes the slot ID
func (a *AttendanceManager) GenerateQRCode(ctx *gin.Context) {
	slotID := ctx.Query("slot_id")
	if slotID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "slot_id required"})
		return
	}

	qrData := map[string]string{
		"slot_id": slotID,
	}
	data, _ := json.Marshal(qrData)

	var png []byte
	png, err := qrcode.Encode(string(data), qrcode.Medium, 256)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR"})
		return
	}

	ctx.Header("Content-Type", "image/png")
	ctx.Writer.Write(png)
}
