package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"storage/configuration"
	"storage/services/user"
	"strings"
)

func AuthMiddleware(conf *configuration.Dependencies) gin.HandlerFunc {
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	return func(c *gin.Context) {
		// Get the token from the "Authorization" header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		// Check if the token has "Bearer " prefix, and strip it
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Parse the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		username := claims["username"]
		var dbUser user.User

		if err := conf.Db.Preload("Roles").Where("lower(username) = ?", username).First(&dbUser).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		var roleNames []string
		for _, role := range dbUser.Roles {
			roleNames = append(roleNames, role.RoleName)
		}

		c.Set("username", username)
		c.Set("user_id", claims["user_id"])
		c.Set("roles", roleNames)

		c.Next()
	}
}

func AllowedRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User is not authenticated"})
			c.Abort()
			return
		}

		print(userRoles)

		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid roles format"})
			c.Abort()
			return
		}

		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			for _, userRole := range roles {
				if strings.ToLower(userRole) == strings.ToLower(allowedRole) {
					roleAllowed = true
					break
				}
			}
			if roleAllowed {
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}
