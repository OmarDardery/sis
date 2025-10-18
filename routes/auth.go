package routes

import (
	"net/http"
	"os"
	"time"

	"github.com/OmarDardery/sis/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func HandleSignup(ctx *gin.Context, db *gorm.DB, role string) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if role == "student" {
		student, err := models.CreateStudent(db, input.Password, input.Email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, student)
		return
	}

	teacher, err := models.CreateTeacher(db, input.Password, input.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, teacher)
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// =====================
// ✅ Signin with JWT
// =====================
func HandleSignin(ctx *gin.Context, db *gorm.DB, role string) {
	var creds models.Credentials
	if err := ctx.ShouldBindJSON(&creds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	var user models.Student
	var teacherUser models.Teacher

	var token *jwt.Token
	if role == "teacher" {

		db.Where("email = ?", creds.Email).First(&teacherUser)
		if teacherUser.Password != creds.Password {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "wrong credentials"})
			return
		}

		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":    teacherUser.ID,
			"email": teacherUser.Email,
			"role":  role,
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		})
	} else {

		db.Where("email = ?", creds.Email).First(&user)
		if user.Password != creds.Password {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "wrong credentials"})
			return
		}
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":    user.ID,
			"email": user.Email,
			"role":  role,
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		})
	}

	// ✅ generate JWT

	tokenStr, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	ctx.JSON(200, gin.H{"token": tokenStr})
}
