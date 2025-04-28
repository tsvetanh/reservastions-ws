package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	"storage/internal/repos"
	. "storage/internal/utils"
)

func HandlerGetAllUsers(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repos.RepoGetAllUsers(conf.Db)
		if err != nil {
			SendError(c, FAILED_GET_USERS, err)
			return
		}
		SendSuccessBody(c, users)
	}
}

func HandlerInsertRole(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		type NewRole struct {
			RoleName string `json:"role_name"`
		}
		var newRole NewRole
		if err := c.BindJSON(&newRole); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		err := repos.RepoInsertRole(conf.Db, newRole.RoleName)
		if err != nil {
			SendError(c, FAILED_CREATE_ROLE, err)
			return
		}
		SendSuccess(c)
	}
}

func HandlerUpdateRole(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		type Role struct {
			RoleId   int64  `json:"role_id"`
			RoleName string `json:"role_name"`
		}
		var role Role
		if err := c.BindJSON(&role); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		err := repos.RepoUpdateRole(conf.Db, role.RoleId, role.RoleName)
		if err != nil {
			SendError(c, FAILED_UPDATE_ROLE, err)
			return
		}
		SendSuccess(c)
	}
}

func HandlerAssignRole(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		type UserRole struct {
			RoleId int64 `json:"role_id"`
			UserId int64 `json:"user_id"`
		}
		var newRole UserRole
		if err := c.BindJSON(&newRole); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		err := repos.RepoAssignRole(conf.Db, newRole.RoleId, newRole.UserId)
		if err != nil {
			SendError(c, FAILED_ASSIGN_ROLE, err)
			return
		}
		SendSuccess(c)
	}
}

func HandlerRevokeRole(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		type UserRole struct {
			RoleId int64 `json:"role_id"`
			UserId int64 `json:"user_id"`
		}
		var newRole UserRole
		if err := c.BindJSON(&newRole); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		err := repos.RepoRevokeRole(conf.Db, newRole.RoleId, newRole.UserId)
		if err != nil {
			SendError(c, FAILED_REVOKE_ROLE, err)
			return
		}
		SendSuccess(c)
	}
}

func HandlerGetAllRoles(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := repos.RepoGetAllRoles(conf.Db)
		if err != nil {
			SendError(c, FAILED_GET_ROLES, err)
			return
		}
		c.JSON(http.StatusOK, roles)
	}
}
