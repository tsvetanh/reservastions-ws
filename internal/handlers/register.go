package handlers

import (
	"storage/configuration"
	"storage/internal/models"
	. "storage/internal/utils"
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
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}
		req.Username = strings.TrimSpace(strings.ToLower(req.Username))

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			SendError(c, FAILED_HASH_PASSWORD, err)
			return
		}

		// Check if the email or username already exists
		var existingUser models.User
		if err = conf.Db.Where("lower(username) = ?", req.Username).First(&existingUser).Error; err == nil {
			SendError(c, USERNAME_EXISTS, err)
			return
		}

		var userRole models.Role
		if err = conf.Db.Where("role_name = ?", "USER").First(&userRole).Error; err != nil {
			userRole.RoleName = "USER"
			if err = conf.Db.Create(&userRole).Error; err != nil {
				SendError(c, FAILED_CREATE_ROLE, err)
			}
			return
		}

		// Create the user model
		u := models.User{
			Username:  req.Username,
			Password:  string(hashedPassword),
			IsActive:  true,
			LastLogin: time.Now(),
		}

		u.Roles = append(u.Roles, userRole)

		// Save the user in the database
		if err := conf.Db.Create(&u).Error; err != nil {
			SendError(c, FAILED_CREATE_USER, err)
			return
		}

		SendSuccess(c)
	}
}
