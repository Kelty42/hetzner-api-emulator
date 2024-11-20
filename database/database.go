package database

import (
	"database/sql"
	"fmt"
	"hetzner-api-emulator/config"
	"hetzner-api-emulator/middlewares"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // MySQL драйвер
	_ "github.com/lib/pq"              // PostgreSQL драйвер
)

// InitDB инициализирует подключение к базе данных
func InitDB(cfg *config.Config) (*sql.DB, int, string, string) {
	var dsn string

	// Формируем строку подключения в зависимости от типа базы данных
	switch cfg.DBType {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	default:
		return nil, http.StatusBadRequest, "UNSUPPORTED_DB_TYPE", fmt.Sprintf("Unsupported database type: %s", cfg.DBType)
	}

	// Подключение к базе данных
	db, err := sql.Open(cfg.DBType, dsn)
	if err != nil {
		// Устанавливаем ошибку через middleware с правильным кодом и статусом
		middlewares.SetError(nil, "DB_CONNECTION_FAILED", http.StatusInternalServerError)
		log.Printf("Failed to open database connection: %v", err)
		return nil, http.StatusInternalServerError, "DB_CONNECTION_FAILED", "Failed to open database connection"
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		// Устанавливаем ошибку через middleware с правильным кодом и статусом
		middlewares.SetError(nil, "DB_PING_FAILED", http.StatusInternalServerError)
		log.Printf("Failed to ping database: %v", err)
		return nil, http.StatusInternalServerError, "DB_PING_FAILED", "Failed to establish a connection to the database"
	}

	log.Println("Database connection established successfully")
	return db, http.StatusOK, "", ""
}
