package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/models"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	switch config.NormalizeDatabaseType(config.DatabaseType) {
	case "sqlite":
		initSQLite()
	default:
		initMySQL()
	}
	if err = migrateAndTune(); err != nil {
		log.Fatal("failed to initialize database: ", err)
	}
}

func initMySQL() {
	dsnCfg, err := mysqlDriver.ParseDSN(config.DatabaseURL)
	if err != nil {
		log.Fatal("failed to parse MySQL DSN: ", err)
	}
	dbName := dsnCfg.DBName
	if dbName == "" {
		log.Fatal("failed to initialize mysql: database name is missing in database_url")
	}

	serverCfg := *dsnCfg
	serverCfg.DBName = ""
	serverDSN := serverCfg.FormatDSN()
	serverDSN = strings.TrimSuffix(serverDSN, "/")
	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		log.Fatal("failed to connect MySQL server: ", err)
	}
	defer db.Close()

	if _, err = db.Exec("CREATE DATABASE IF NOT EXISTS `" + strings.ReplaceAll(dbName, "`", "``") + "`"); err != nil {
		log.Fatal("failed to create database: ", err)
	}

	DB, err = gorm.Open(mysql.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	log.Println("MySQL database connected successfully")
}

func initSQLite() {
	dbPath := config.DatabasePath
	if dbPath == "" {
		dbPath = "data/kards.db"
	}
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal("failed to create sqlite directory: ", err)
		}
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect sqlite database: ", err)
	}
	log.Printf("SQLite database connected successfully: %s", dbPath)
}

func migrateAndTune() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err = DB.AutoMigrate(&models.User{}, &models.Deck{}); err != nil {
		return err
	}

	ensureIndexes()
	log.Printf("%s database initialized successfully", strings.ToUpper(config.NormalizeDatabaseType(config.DatabaseType)))
	return nil
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
