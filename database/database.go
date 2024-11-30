package database

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"hetzner-api-emulator/middlewares"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func ConnectDatabase() {
	dbConnection := os.Getenv("DB_CONNECTION")
	var dsn string
	var dialector gorm.Dialector

	switch dbConnection {
	case "mysql":
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)
		dialector = mysql.Open(dsn)

	case "postgres":
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		dialector = postgres.Open(dsn)

	default:
		middlewares.SetError(nil, "DB_UNSUPPORTED_TYPE", http.StatusInternalServerError)
		panic("Invalid DB_CONNECTION configuration")
	}

	var err error
	db, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		middlewares.SetError(nil, "Failed to connect to database", http.StatusInternalServerError)
		panic("Database connection error")
	}

	log.Println("Database connection established successfully")
}

// GetDB returns the active database connection
func GetDB() *gorm.DB {
	return db
}
