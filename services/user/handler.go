package user

import (
	"net/http"
	"storage/configuration"

	"github.com/gin-gonic/gin"
)

func HandlerGetAllUsers(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := RepoGetAllUsers(conf.Db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func HandlerInsertRole(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		type NewRole struct {
			RoleName string `json:"role_name"`
		}
		var newRole NewRole
		if err := c.BindJSON(&newRole); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		err := RepoInsertRole(conf.Db, newRole.RoleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save role: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully created role"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		err := RepoUpdateRole(conf.Db, role.RoleId, role.RoleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully updated role"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		err := RepoAssignRole(conf.Db, newRole.RoleId, newRole.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully assigned role"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		err := RepoRevokeRole(conf.Db, newRole.RoleId, newRole.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke role: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully revoked role"})
	}
}
