package config

import (
	"fmt"
	"go-chat/internal/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	cfg := AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Logger.Fatal("数据库连接失败: ", zap.Error(err))
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	logger.Logger.Info("数据库连接成功")

}
