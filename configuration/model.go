package configuration

import (
	"gorm.io/gorm"
)

type Dependencies struct {
	Cfg *EnvironmentConfig
	Db  *gorm.DB
}

type MainConfig struct {
	ActiveEnvironment string              `json:"active_env"`
	EnvironmentConfig []EnvironmentConfig `json:"env_config" validate:"required,dive"`
}

type EnvironmentConfig struct {
	EnvType  string   `json:"env_type" validate:"required"`
	Port     string   `json:"port" validate:"required,min=2,max=5,numeric"`
	Database Database `json:"database" validate:"required"`
}

type Database struct {
	User         string `json:"user"`
	Password     string `json:"password"`
	DatabaseName string `json:"database_name"`
	Port         string `json:"port"`
	Host         string `json:"host"`
}

//type DB struct {
//	Server          string        `json:"server" validate:"required,hostname|ip4_addr"`
//	Port            string        `json:"port" validate:"required,min=2,max=5,numeric"`
//	Service         string        `json:"service" validate:"required,max=50"`
//	MaxOpenConns    int           `json:"max_open_conns" validate:"required,gte=1,lte=500"`
//	MaxIdleConns    int           `json:"max_idle_conns" validate:"required,gte=1,lte=450"`
//	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" validate:"required"`
//}
