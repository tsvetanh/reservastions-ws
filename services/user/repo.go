package user

import "gorm.io/gorm"

func RepoGetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	if err := db.Preload("Roles").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
func RepoInsertRole(db *gorm.DB, roleName string) error {
	if err := db.Create(&Role{RoleName: roleName}).Error; err != nil {
		return err
	}
	return nil
}

func RepoUpdateRole(db *gorm.DB, roleId int64, roleName string) error {
	if err := db.Model(&Role{}).Where("id = ?", roleId).Update("role_name", roleName).Error; err != nil {
		return err
	}
	return nil
}
func RepoAssignRole(db *gorm.DB, roleId, userId int64) error {
	if err := db.Create(&UserRoles{RoleID: roleId, UserID: userId}).Error; err != nil {
		return err
	}
	return nil
}

func RepoRevokeRole(db *gorm.DB, roleId, userId int64) error {
	if err := db.Delete(&UserRoles{RoleID: roleId, UserID: userId}).Error; err != nil {
		return err
	}
	return nil
}
