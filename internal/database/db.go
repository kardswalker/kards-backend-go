package database

import (
	"database/sql"
	"log"
	"strings"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/models"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	serverDSN := strings.Replace(config.DatabaseURL, "/users?", "/?", 1)
	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		log.Fatal("failed to connect MySQL server: ", err)
	}
	defer db.Close()

	if _, err = db.Exec("CREATE DATABASE IF NOT EXISTS users"); err != nil {
		log.Fatal("failed to create database: ", err)
	}

	DB, err = gorm.Open(mysql.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	if err = DB.AutoMigrate(&models.User{}, &models.Deck{}); err != nil {
		log.Fatal("failed to migrate database schema: ", err)
	}

	ensureIndexes()
	log.Println("MySQL database initialized successfully")
}

func ensureIndexes() {
	ensureIndex(&models.User{}, "Username")
	ensureIndex(&models.Deck{}, "UserID")
}

func ensureIndex(model interface{}, name string) {
	if DB.Migrator().HasIndex(model, name) {
		return
	}

	if err := DB.Migrator().CreateIndex(model, name); err != nil {
		log.Printf("failed to create index %s: %v", name, err)
	}
}
