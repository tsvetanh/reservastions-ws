package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"storage/configuration"
	. "storage/internal/models"
	"storage/internal/repository"
	. "storage/internal/utils"
	"strconv"
)

type UserHandler struct {
	UserRepo *repository.BaseRepository[User]
	RoleRepo *repository.BaseRepository[Role]
	Conf     *configuration.Dependencies
}

func NewUserHandler(conf *configuration.Dependencies) *UserHandler {
	return &UserHandler{
		UserRepo: &repository.BaseRepository[User]{DB: conf.Db},
		RoleRepo: &repository.BaseRepository[Role]{DB: conf.Db},
		Conf:     conf,
	}
}

// GetAllUsers retrieves all users
func (h *UserHandler) GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []User
		if err := h.UserRepo.GetAll(&users); err != nil {
			SendError(c, FAILED_GET_USERS, err)
			return
		}
		SendSuccessBody(c, users)
	}
}

// GetAllRoles retrieves all roles
func (h *UserHandler) GetAllRoles() gin.HandlerFunc {
	return func(c *gin.Context) {
		var roles []Role
		if err := h.RoleRepo.GetAll(&roles); err != nil {
			SendError(c, FAILED_GET_ROLES, err)
			return
		}
		SendSuccessBody(c, roles)
	}
}

// InsertRole handles the creating of new roles
func (h *UserHandler) InsertRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		var newRole Role
		if err := c.ShouldBindJSON(&newRole); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}
		if err := h.RoleRepo.Create(&newRole); err != nil {
			SendError(c, FAILED_CREATE_ROLE, err)
			return
		}
		SendSuccess(c)
	}
}

// UpdateRole updates a role by id
func (h *UserHandler) UpdateRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		type input struct {
			RoleName string `json:"role_name" binding:"required"`
		}
		roleIDParam := c.Param("id")
		roleID, err := strconv.ParseInt(roleIDParam, 10, 64)
		if err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, fmt.Errorf("invalid role ID"))
			return
		}

		var dto input
		if err := c.ShouldBindJSON(&dto); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		role := Role{
			RoleID:   roleID,
			RoleName: dto.RoleName,
		}

		if err := h.RoleRepo.Update(roleID, &role); err != nil {
			SendError(c, FAILED_UPDATE_ROLE, err)
			return
		}

		SendSuccess(c)
	}
}

func (h *UserHandler) AssignRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		var input struct {
			RoleID int64 `json:"role_id"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			SendError(c, INVALID_USER_ID_PARAM, err)
			return
		}

		var user User
		if err := h.Conf.Db.Preload("Roles").First(&user, userID).Error; err != nil {
			SendError(c, USER_NOT_FOUND, err)
			return
		}

		var role Role
		if err := h.Conf.Db.First(&role, input.RoleID).Error; err != nil {
			SendError(c, ROLE_NOT_FOUND, err)
			return
		}

		if err := h.Conf.Db.Model(&user).Association("Roles").Append(&role); err != nil {
			SendError(c, FAILED_ASSIGN_ROLE, err)
			return
		}

		SendSuccess(c)
	}
}

func (h *UserHandler) RevokeRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		roleIDStr := c.Param("role")

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			SendError(c, INVALID_USER_ID_PARAM, err)
			return
		}
		roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
		if err != nil {
			SendError(c, INVALID_ROLE_ID_PARAM, err)
			return
		}

		var user User
		if err := h.Conf.Db.Preload("Roles").First(&user, userID).Error; err != nil {
			SendError(c, USER_NOT_FOUND, err)
			return
		}

		var role Role
		if err := h.Conf.Db.First(&role, roleID).Error; err != nil {
			SendError(c, ROLE_NOT_FOUND, err)
			return
		}

		if err := h.Conf.Db.Model(&user).Association("Roles").Delete(&role); err != nil {
			SendError(c, FAILED_REVOKE_ROLE, err)
			return
		}

		SendSuccess(c)
	}
}
