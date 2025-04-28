package handlers

import (
	"os"
	"storage/configuration"
	"storage/internal/models"
	. "storage/internal/utils"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler handles the login requests
func LoginHandler(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		type User struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var inputUser User
		jwtKey := os.Getenv("JWT_SECRET_KEY")

		if err := c.ShouldBindJSON(&inputUser); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		inputUser.Username = strings.TrimSpace(strings.ToLower(inputUser.Username))

		var dbUser models.User
		if err := conf.Db.Preload("Roles").Where("lower(username) = ?", inputUser.Username).First(&dbUser).Error; err != nil {
			SendError(c, INVALID_CREDENTIALS, err)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(inputUser.Password)); err != nil {
			SendError(c, INVALID_CREDENTIALS, err)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": inputUser.Username,
			"exp":      time.Now().Add(time.Hour * 1).Unix(),
		})

		tokenString, err := token.SignedString([]byte(jwtKey))
		if err != nil {
			SendError(c, TOKEN_CREATION_FAILED, err)
			return
		}

		SendSuccessBody(c, gin.H{
			"username": inputUser.Username,
			"token":    tokenString,
			"roles":    dbUser.GetRoleNames(),
		})
	}
}
