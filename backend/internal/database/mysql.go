package database

import (
	"fmt"
	"log"

	"book-trading/backend/internal/config"
	"book-trading/backend/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitMySQL() {
	cfg := config.AppConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL: ", err)
	}

	// 自动迁移 - 根据模型结构创建或更新表
	// 这会自动创建 users 表（如果不存在）
	err = DB.AutoMigrate(&models.User{}, &models.Batch{}, &models.Message{}, &models.Conversation{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	log.Println("MySQL connected and migrated successfully")
}
