package database

import (
	"kards-backend-go/internal/models"
	"log"

	"github.com/glebarez/sqlite" 
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ 数据库连接失败: ", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.Deck{})
	if err != nil {
		log.Fatal("❌ 表结构迁移失败: ", err)
	}
	log.Println("✅ SQLite 数据库初始化成功 (文件: gorm.db)")
}
