package main

import (
	"log"
	"net/http"
	"os"

	"github.com/OmarDardery/sis/middleware"
	"github.com/OmarDardery/sis/models"
	"github.com/OmarDardery/sis/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize database
	db := models.Migrate()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get sql.DB from gorm.DB:", err)
	}
	defer sqlDB.Close()

	// Initialize Gin
	server := gin.Default()

	// Serve static files (JS, CSS, etc.)
	server.Static("/static", "./frontend/static")

	// Load HTML templates
	server.LoadHTMLGlob("./frontend/*.html")

	// ================= FRONTEND ROUTES =================
	server.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})
	server.GET("/signin", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "signin.html", gin.H{
			"API_BASE":            os.Getenv("API_BASE"),
			"FRONTEND_SECRET_KEY": os.Getenv("FRONTEND_SECRET_KEY"),
		})
	})

	server.GET("/signup", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "signup.html", gin.H{
			"API_BASE":            os.Getenv("API_BASE"),
			"FRONTEND_SECRET_KEY": os.Getenv("FRONTEND_SECRET_KEY"),
		})
	})

	server.GET("/student/home", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "student_home.html", gin.H{
			"FRONTEND_SECRET_KEY": os.Getenv("FRONTEND_SECRET_KEY"),
		})
	})

	server.GET("/teacher/home", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "teacher_home.html", gin.H{
			"FRONTEND_SECRET_KEY": os.Getenv("FRONTEND_SECRET_KEY"),
		})
	})

	// ================= AUTH ROUTES =================
	server.POST("/signup/student", func(ctx *gin.Context) {
		routes.HandleSignup(ctx, db, "student")
	})
	server.POST("/signup/teacher", func(ctx *gin.Context) {
		routes.HandleSignup(ctx, db, "teacher")
	})
	server.POST("/signin/student", func(ctx *gin.Context) {
		routes.HandleSignin(ctx, db, "student")
	})
	server.POST("/signin/teacher", func(ctx *gin.Context) {
		routes.HandleSignin(ctx, db, "teacher")
	})

	// ================= MODEL CREATION ROUTES =================
	server.POST("/subjects", func(ctx *gin.Context) {
		routes.PostSubject(ctx, db)
	})
	server.POST("/semesters", func(ctx *gin.Context) {
		routes.PostSemester(ctx, db)
	})
	server.POST("/slots", func(ctx *gin.Context) {
		routes.PostSlots(ctx, db)
	})

	// ================= ATTENDANCE (WebSocket + API) =================
	attendanceManager := &routes.AttendanceManager{
		Attendances: make(map[uint][]models.Attendance),
	}

	server.GET("/ws/attendance", func(ctx *gin.Context) {
		attendanceManager.WebsocketAttendance(ctx, db)
	})

	server.POST("/attendance/mark", middleware.StudentAuthMiddleware(), func(ctx *gin.Context) {
		attendanceManager.MarkStudentAttendance(ctx, db)
	})

	server.GET("/attendance/qr", func(ctx *gin.Context) {
		attendanceManager.GenerateQRCode(ctx)
	})
	server.GET("/attendance", func(ctx *gin.Context) {
		var as []models.Attendance
		db.Preload("Slot").Preload("Slot.Teacher").Preload("Slot.Semester").Preload("Slot.Subject").Preload("Student").Find(&as)
		ctx.JSON(200, gin.H{
			"data": as,
		})
	})
	// ================= START SERVER =================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("🚀 Server running on http://localhost:%s", port)
	server.Run("0.0.0.0:" + port)
}
