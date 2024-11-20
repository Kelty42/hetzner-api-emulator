package middlewares

import (
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"github.com/gin-gonic/gin"
)

// DBAuthMiddleware проверяет базовую авторизацию с использованием базы данных
func DBAuthMiddleware(db *sql.DB, dbType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем значение авторизации из заголовка
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Authorization header is missing")
			SetError(c, "AUTH_HEADER_MISSING", http.StatusUnauthorized)
			RespondWithError(c, http.StatusUnauthorized, "AUTH_HEADER_MISSING", "Authorization header required")
			c.Abort()
			return
		}

		// Проверяем формат заголовка (Basic ...)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			log.Println("Invalid authorization format")
			SetError(c, "INVALID_AUTH_FORMAT", http.StatusUnauthorized)
			RespondWithError(c, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Invalid authorization format")
			c.Abort()
			return
		}

		// Декодируем Base64
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			log.Printf("Base64 decoding error: %v", err)
			SetError(c, "INVALID_BASE64_ENCODING", http.StatusUnauthorized)
			RespondWithError(c, http.StatusUnauthorized, "INVALID_BASE64_ENCODING", "Invalid Base64 encoding")
			c.Abort()
			return
		}

		// Разделяем username и password
		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			log.Println("Invalid username or password format")
			SetError(c, "INVALID_USERNAME_PASSWORD_FORMAT", http.StatusUnauthorized)
			RespondWithError(c, http.StatusUnauthorized, "INVALID_USERNAME_PASSWORD_FORMAT", "Invalid username or password format")
			c.Abort()
			return
		}
		username := credentials[0]
		password := strings.TrimSpace(credentials[1]) // Убираем пробелы из пароля

		// Формируем запрос в зависимости от типа базы данных
		var query string
		switch dbType {
		case "mysql":
			query = "SELECT id, password FROM users WHERE username = ?"
		case "postgres":
			query = "SELECT id, password FROM users WHERE username = $1"
		default:
			log.Println("Unsupported database type")
			SetError(c, "UNSUPPORTED_DB_TYPE", http.StatusInternalServerError)
			RespondWithError(c, http.StatusInternalServerError, "UNSUPPORTED_DB_TYPE", "Unsupported database type")
			c.Abort()
			return
		}

		// Выполняем запрос к базе данных
		var userID int
		var storedPasswordHash string
		err = db.QueryRow(query, username).Scan(&userID, &storedPasswordHash)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Println("User not found in the database")
				SetError(c, "INVALID_USERNAME_PASSWORD", http.StatusUnauthorized)
				RespondWithError(c, http.StatusUnauthorized, "INVALID_USERNAME_PASSWORD", "Invalid username or password")
				c.Abort()
				return
			}
			log.Printf("Database query error: %v", err)
			SetError(c, "DB_QUERY_ERROR", http.StatusInternalServerError)
			RespondWithError(c, http.StatusInternalServerError, "DB_QUERY_ERROR", "Database error")
			c.Abort()
			return
		}

		// Проверяем пароль
		err = bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(password))
		if err != nil {
			log.Printf("Password verification failed: %v", err)
			SetError(c, "INVALID_USERNAME_PASSWORD", http.StatusUnauthorized)
			RespondWithError(c, http.StatusUnauthorized, "INVALID_USERNAME_PASSWORD", "Invalid username or password")
			c.Abort()
			return
		}

		// Аутентификация прошла успешно, сохраняем user_id в контексте
		log.Printf("Authentication successful for user: %s", username)
		c.Set("user_id", userID) // Добавляем user_id в контекст
		c.Next()
	}
}
