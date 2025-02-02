package register

import (
	"net/http"
	"storage/configuration"
	"storage/services/user"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserModel represents the user model in the database
type UserModel struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

// RegisterRequest represents the expected request body for registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterHandler handles user registration
func RegisterHandler(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest

		// Bind and validate the JSON payload
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		req.Username = strings.TrimSpace(strings.ToLower(req.Username))

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Check if the email or username already exists
		var existingUser user.User
		if err := conf.Db.Where("lower(username) = ?", req.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}

		// Create the user model
		user := user.User{
			Username:  req.Username,
			Password:  string(hashedPassword),
			IsActive:  true, // Set default values as necessary
			LastLogin: time.Now(),
		}

		// Save the user in the database
		if err := conf.Db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}
