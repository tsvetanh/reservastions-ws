package repos

import (
	"gorm.io/gorm"
	"storage/internal/models"
)

func RepoGetAllUsers(db *gorm.DB) ([]models.User, error) {
	var users []models.User
	if err := db.Preload("Roles").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
func RepoInsertRole(db *gorm.DB, roleName string) error {
	if err := db.Create(&models.Role{RoleName: roleName}).Error; err != nil {
		return err
	}
	return nil
}

func RepoUpdateRole(db *gorm.DB, roleId int64, roleName string) error {
	if err := db.Model(&models.Role{}).Where("id = ?", roleId).Update("role_name", roleName).Error; err != nil {
		return err
	}
	return nil
}
func RepoAssignRole(db *gorm.DB, roleId, userId int64) error {
	if err := db.Create(&models.UserRoles{RoleID: roleId, UserID: userId}).Error; err != nil {
		return err
	}
	return nil
}

func RepoRevokeRole(db *gorm.DB, roleId, userId int64) error {
	if err := db.Delete(&models.UserRoles{RoleID: roleId, UserID: userId}).Error; err != nil {
		return err
	}
	return nil
}

func RepoGetAllRoles(db *gorm.DB) ([]models.Role, error) {
	var roles []models.Role
	if err := db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
