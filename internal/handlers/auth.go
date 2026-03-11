package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"storage/configuration"
	. "storage/internal/models"
	"storage/internal/repository"
	. "storage/internal/utils"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UserRepo         *repository.BaseRepository[User]
	RoleRepo         *repository.BaseRepository[Role]
	RefreshTokenRepo *repository.BaseRepository[RefreshToken]
	Conf             *configuration.Dependencies
	Env              string
}

func NewAuthHandler(conf *configuration.Dependencies) *AuthHandler {
	return &AuthHandler{
		UserRepo:         &repository.BaseRepository[User]{DB: conf.Db},
		RoleRepo:         &repository.BaseRepository[Role]{DB: conf.Db},
		RefreshTokenRepo: &repository.BaseRepository[RefreshToken]{DB: conf.Db},
		Conf:             conf,
		Env:              conf.Cfg.EnvType,
	}
}

// RegisterHandler handles user registration
func (h *AuthHandler) RegisterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}
		req.Username = strings.TrimSpace(strings.ToLower(req.Username))

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			SendError(c, FAILED_HASH_PASSWORD, err)
			return
		}

		var existingUser User
		if err = h.UserRepo.DB.Where("lower(username) = ?", req.Username).First(&existingUser).Error; err == nil {
			SendError(c, USERNAME_EXISTS, err)
			return
		}

		var userRole Role
		if err = h.RoleRepo.DB.Where("role_name = ?", "USER").First(&userRole).Error; err != nil {
			userRole.RoleName = "USER"
			if err = h.RoleRepo.Create(&userRole); err != nil {
				SendError(c, FAILED_CREATE_ROLE, err)
			}
			return
		}

		u := User{
			Username:  req.Username,
			Password:  string(hashedPassword),
			IsActive:  true,
			LastLogin: time.Now(),
		}

		u.Roles = append(u.Roles, userRole)

		// Save the user in the database
		if err := h.UserRepo.Create(&u); err != nil {
			SendError(c, FAILED_CREATE_USER, err)
			return
		}

		SendSuccess(c)
	}
}

// LoginHandler handles the login requests
func (h *AuthHandler) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		type inUser struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var inputUser inUser
		jwtKey := os.Getenv("JWT_SECRET_KEY")
		//env := h.Env

		if err := c.ShouldBindJSON(&inputUser); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		inputUser.Username = strings.TrimSpace(strings.ToLower(inputUser.Username))

		var dbUser User
		if err := h.UserRepo.DB.Preload("Roles").Where("lower(username) = ?", inputUser.Username).First(&dbUser).Error; err != nil {
			SendError(c, INVALID_CREDENTIALS, err)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(inputUser.Password)); err != nil {
			SendError(c, INVALID_CREDENTIALS, err)
			return
		}

		// access token
		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": dbUser.Username,
			"roles":    dbUser.Roles,
			"exp":      time.Now().Add(time.Hour * 1).Unix(),
		})

		accessTokenString, err := accessToken.SignedString([]byte(jwtKey))
		if err != nil {
			SendError(c, TOKEN_CREATION_FAILED, err)
			return
		}

		// refresh token
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": dbUser.Username,
			"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
		})
		refreshTokenString, err := refreshToken.SignedString([]byte(jwtKey))
		if err != nil {
			SendError(c, TOKEN_CREATION_FAILED, err)
			return
		}

		hash := sha256.Sum256([]byte(refreshTokenString))
		rt := RefreshToken{
			TokenHash: hex.EncodeToString(hash[:]),
			UserID:    dbUser.UserID,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		if err := h.RefreshTokenRepo.Create(&rt); err != nil {
			SendError(c, FAILED_CREATE_TOKEN, err)
			return
		}
		log.Printf("Setting access_token cookie: %s", accessTokenString)
		//secure := env == "prod"
		c.SetCookie(
			"access_token",
			accessTokenString,
			15*60,
			"/",
			"",
			false,
			true,
		)
		h.UserRepo.DB.Model(&dbUser).Update("last_login", time.Now())

		SendSuccessBody(c, gin.H{
			"username": dbUser.Username,
			"roles":    dbUser.GetRoleNames(),
			"token":    accessTokenString,
		})
	}
}

func (h *AuthHandler) RefreshHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtKey := os.Getenv("JWT_SECRET_KEY")

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			SendError(c, MISSING_TOKEN, err)
			return
		}

		token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})
		if err != nil || !token.Valid {
			SendError(c, INVALID_TOKEN, err)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			SendError(c, INVALID_TOKEN_CLAIMS, nil)
			return
		}

		hash := sha256.Sum256([]byte(refreshToken))
		var dbToken RefreshToken
		if err := h.RefreshTokenRepo.DB.Where("token_hash = ?", hex.EncodeToString(hash[:])).First(&dbToken).Error; err != nil {
			SendError(c, INVALID_TOKEN, err)
			return
		}

		if time.Now().After(dbToken.ExpiresAt) {
			h.RefreshTokenRepo.Delete(&dbToken)
			SendError(c, INVALID_TOKEN, err)
			return
		}

		userID := int64(claims["user_id"].(float64))
		var dbUser User
		if err := h.RefreshTokenRepo.DB.Preload("Roles").First(&dbUser, userID).Error; err != nil {
			SendError(c, TOKEN_CREATION_FAILED, err)
			return
		}

		newAccess := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  dbUser.UserID,
			"username": dbUser.Username,
			"roles":    dbUser.GetRoleNames(),
			"exp":      time.Now().Add(15 * time.Minute).Unix(),
		})
		newAccessStr, _ := newAccess.SignedString([]byte(jwtKey))

		SendSuccessBody(c, gin.H{
			"token": newAccessStr,
		})
	}
}

func (h *AuthHandler) LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no refresh token"})
			SendError(c, MISSING_TOKEN, err)
			return
		}

		hash := sha256.Sum256([]byte(refreshToken))
		h.RefreshTokenRepo.DB.Where("token_hash = ?", hex.EncodeToString(hash[:])).Delete(&RefreshToken{})

		env := h.Env
		secure := env == "prod"
		c.SetCookie("refresh_token", "", -1, "/", "", secure, true)

		SendSuccess(c)
	}
}
