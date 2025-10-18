package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/OmarDardery/sis/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type AttendanceManager struct {
	Mu          sync.Mutex
	Attendances map[uint][]models.Attendance // slot_id → list of student attendances
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	TeacherToken string `json:"teacher_token"`
	SlotID       uint   `json:"slot_id"`
}

type StudentAttendanceInput struct {
	SlotID    uint   `json:"slot_id"`
	SecretKey string `json:"secret_key"`
}

// =========================================================
// WebSocket: teacher opens attendance session for a slot
// =========================================================
func (a *AttendanceManager) WebsocketAttendance(ctx *gin.Context, db *gorm.DB) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot establish websocket connection"})
		return
	}

	defer conn.Close()

	// Inside WebsocketAttendance, after reading the message:
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		conn.WriteJSON(gin.H{"error": "invalid message format"})
		return
	}

	// === Parse teacher JWT ===
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(msg.TeacherToken, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure it's signed with HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		conn.WriteJSON(gin.H{"error": "invalid token"})
		return
	}

	// Extract teacher ID from claims
	teacherIDFloat, ok := claims["id"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "invalid token claims"})
		return
	}
	teacherID := uint(teacherIDFloat)

	// === Check if teacher teaches this slot ===
	var slot models.Slot
	if err := db.First(&slot, msg.SlotID).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "slot not found"})
		return
	}

	if slot.TeacherID != teacherID {
		conn.WriteJSON(gin.H{"error": "unauthorized"})
		return
	}

	// === Open attendance session ===
	a.Mu.Lock()
	a.Attendances[msg.SlotID] = []models.Attendance{}
	a.Mu.Unlock()

	// === Send back a QR link (frontend will convert to QR image) ===
	qrData := fmt.Sprintf(`{"slot_id": %d}`, msg.SlotID)
	conn.WriteJSON(gin.H{
		"message": "attendance session opened",
		"slot_id": msg.SlotID,
		"qr_data": qrData,
	})
	log.Printf("📘 Attendance session opened for slot %d", msg.SlotID)

	// Wait for disconnection
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("🔌 Teacher disconnected from slot %d — saving attendance", msg.SlotID)
			a.closeAttendanceSession(db, msg.SlotID)
			break
		}
	}
}

// =========================================================
// HTTP: student marks attendance (JWT + secret key protected)
// =========================================================
func (a *AttendanceManager) MarkStudentAttendance(ctx *gin.Context, db *gorm.DB) {
	var input StudentAttendanceInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// === Validate secret key ===
	secret := os.Getenv("FRONTEND_SECRET_KEY")
	if secret == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfiguration"})
		return
	}
	if input.SecretKey != secret {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "invalid secret key"})
		return
	}

	// === Get JWT from Authorization header ===
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	// === Validate JWT ===
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	studentIDFloat, ok := claims["id"].(float64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}
	studentID := uint(studentIDFloat)

	// === Record attendance ===
	a.Mu.Lock()
	defer a.Mu.Unlock()

	records, exists := a.Attendances[input.SlotID]
	if !exists {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "attendance not open for this slot"})
		return
	}

	for _, r := range records {
		if r.StudentID == studentID {
			ctx.JSON(http.StatusConflict, gin.H{"error": "student already marked present"})
			return
		}
	}

	newRecord := models.Attendance{
		SlotID:    input.SlotID,
		StudentID: studentID,
		Status:    "present",
	}

	a.Attendances[input.SlotID] = append(records, newRecord)
	ctx.JSON(http.StatusOK, gin.H{
		"message":    "attendance marked successfully",
		"student_id": studentID,
		"slot_id":    input.SlotID,
	})
}

// =========================================================
// Helper: save and clear slot attendance when teacher disconnects
// =========================================================
func (a *AttendanceManager) closeAttendanceSession(db *gorm.DB, slotID uint) {
	a.Mu.Lock()
	defer a.Mu.Unlock()

	records, exists := a.Attendances[slotID]
	if !exists {
		return
	}

	if len(records) > 0 {
		if err := db.Create(&records).Error; err != nil {
			log.Printf("❌ Failed to save attendance for slot %d: %v", slotID, err)
		} else {
			log.Printf("✅ Saved %d attendance records for slot %d", len(records), slotID)
		}
	}

	delete(a.Attendances, slotID)
}
