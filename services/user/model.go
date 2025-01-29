package user

import (
	"time"
)

type User struct {
	UserID    int64     `gorm:"column:id;primaryKey" json:"id"`
	Username  string    `gorm:"column:username;size:50;unique;not null" json:"username"`
	Email     string    `gorm:"column:email;size:100;not null;unique" json:"email"`
	Password  string    `gorm:"column:password;size:255;not null" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	LastLogin time.Time `gorm:"column:last_login" json:"last_login"`
	IsActive  bool      `gorm:"column:is_active" json:"is_active"`
	Roles     []Role    `gorm:"many2many:hall_res_project.users_roles;joinForeignKey:UserID;joinReferences:RoleID" json:"roles,omitempty"`
}

type UserRoles struct {
	RoleID int64 `gorm:"column:role_id;primaryKey"`
	UserID int64 `gorm:"column:user_id;primaryKey"`
}

type Role struct {
	RoleID    int64     `gorm:"column:id;primaryKey" json:"role_id"`
	RoleName  string    `gorm:"column:role_name;unique;size:255;not null" json:"role_name"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	Users     []User    `gorm:"many2many:hall_res_project.users_roles;joinForeignKey:RoleID;joinReferences:UserID" json:"users,omitempty"`
}

type ChangePassword struct {
	Username       string `json:"username"`
	CurrPassword   string `json:"curr_password"`
	NewPassword    string `json:"new_password"`
	RepeatPassword string `json:"repeat_password"`
}

func (Role) TableName() string {
	return "hall_res_project.roles"
}

func (UserRoles) TableName() string {
	return "hall_res_project.users_roles"
}

func (User) TableName() string {
	return "hall_res_project.users"
}
