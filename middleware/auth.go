package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"storage/configuration"
	"storage/internal/models"
	. "storage/internal/utils"
	"strings"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type UserDetails struct {
	Username string
	Roles    []string
}

func (u UserDetails) HasRole(role string) bool {
	role = strings.ToLower(role)
	for _, r := range u.Roles {
		if strings.ToLower(r) == role {
			return true
		}
	}
	return false
}

func AuthMiddleware(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString == "" {
			SendError(c, MISSING_TOKEN, nil)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			SendError(c, INVALID_TOKEN, err)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			SendError(c, INVALID_TOKEN_CLAIMS, nil)
			return
		}
		username, ok := claims["username"].(string)
		if !ok {
			SendError(c, USERNAME_NOT_IN_TOKEN, nil)
			return
		}

		var dbUser models.User
		if err := conf.Db.Preload("Roles").Where("lower(username) = ?", strings.ToLower(username)).First(&dbUser).Error; err != nil {
			SendError(c, INVALID_CREDENTIALS, err)
			return
		}

		c.Set("user", UserDetails{
			Username: dbUser.Username,
			Roles:    dbUser.GetRoleNames(),
		})

		c.Next()
	}
}

// AllowedRoles checks if the user has the given roles and lets the request through if not returns 401
func AllowedRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		u, ok := c.Get("user")
		if !ok {
			SendError(c, USER_NOT_IN_CONTEXT, nil)
			return
		}
		roles := u.(UserDetails).Roles

		for _, allowedRole := range allowedRoles {
			for _, userRole := range roles {
				if strings.ToLower(userRole) == strings.ToLower(allowedRole) || userRole == "ADMIN" {
					c.Next()
					return
				}
			}
		}

		SendError(c, ACCESS_DENIED, nil)
	}
}
