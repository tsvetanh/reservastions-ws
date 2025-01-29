package configuration

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"time"
)

func Init() (*Dependencies, error) {
	cfg, err := loadCfg()
	if err != nil {
		return nil, err
	}

	gin.SetMode(gin.DebugMode)

	var usedConfig EnvironmentConfig

	for _, env := range cfg.EnvironmentConfig {
		if env.EnvType == cfg.ActiveEnvironment {
			usedConfig = env
			break
		}
	}

	if cfg.ActiveEnvironment == "PROD" {
		//gin.SetMode(gin.ReleaseMode)
	}

	db, err := connectDb(&usedConfig.Database)
	if err != nil {
		return nil, err
	}

	return &Dependencies{
		Cfg: &usedConfig,
		Db:  db,
	}, nil
}

func KeepConnectionsAlive(db *gorm.DB, interval time.Duration) {
	for {
		db.Exec("SELECT 1")
		time.Sleep(interval)
	}
}

func loadCfg() (*MainConfig, error) {
	file, err := os.Open("./configuration/config.json")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var cfg MainConfig

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func connectDb(cfg *Database) (*gorm.DB, error) {
	dsn := cfg.User + "@tcp(" + cfg.Host + ":" + cfg.Port + ")/" + cfg.DatabaseName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func validateConfig(cfg *MainConfig) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return err
	}
	return nil
}
