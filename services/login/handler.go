package login

import (
	"net/http"
	"os"
	"storage/configuration"
	"storage/services/user"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User struct to represent the expected request body
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims struct to include within the JWT token
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginHandler handles the login requests
func LoginHandler(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var inputUser User
		jwtKey := os.Getenv("JWT_SECRET_KEY")
		if err := c.ShouldBindJSON(&inputUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		inputUser.Username = strings.TrimSpace(strings.ToLower(inputUser.Username))

		var dbUser user.User // Use the User model from the users package
		if err := conf.Db.Preload("Roles").Where("lower(username) = ?", inputUser.Username).First(&dbUser).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(inputUser.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": inputUser.Username,
			"exp":      time.Now().Add(time.Hour * 1).Unix(),
		})

		tokenString, err := token.SignedString([]byte(jwtKey))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"username": inputUser.Username,
			"token":    tokenString,
			"roles":    dbUser.Roles,
		})
	}
}
