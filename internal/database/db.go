package database

import (
	"database/sql"
	"kards-backend-go/internal/config"
	"kards-backend-go/internal/models"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// 先连接到MySQL服务器（不指定数据库）以创建数据库
	// 从config.DatabaseURL中提取服务器DSN
	serverDSN := strings.Replace(config.DatabaseURL, "/users?", "/?", 1)
	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		log.Fatal("❌ 连接MySQL服务器失败: ", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS users")
	if err != nil {
		log.Fatal("❌ 创建数据库失败: ", err)
	}

	// 现在连接到users数据库
	DB, err = gorm.Open(mysql.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ 数据库连接失败: ", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.Deck{})
	if err != nil {
		log.Fatal("❌ 表结构迁移失败: ", err)
	}
	log.Println("✅ MySQL 数据库初始化成功")
}
