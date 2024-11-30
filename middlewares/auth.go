package middlewares

import (
	"encoding/base64"
	"hetzner-api-emulator/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DBAuthMiddleware проверяет базовую авторизацию с использованием базы данных
func DBAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем авторизацию для маршрута регистрации
		if c.Request.URL.Path == "/register" {
			c.Next()
			return
		}

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
		password := strings.TrimSpace(credentials[1])

		// Логируем полученные данные для отладки
		log.Printf("Decoded credentials: username=%s, password=%s", username, password)

		// Поиск пользователя в базе данных
		var user models.User

		err = db.Where("username = ?", username).First(&user).Error
		if err != nil {
			log.Println("User not found or database error")
			SetError(c, "INVALID_USERNAME_PASSWORD", http.StatusUnauthorized)
			RespondWithError(c, http.StatusUnauthorized, "INVALID_USERNAME_PASSWORD", "Invalid username or password")
			c.Abort()
			return
		}

		// Проверяем пароль
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			log.Printf("Password verification failed: expected hash=%s, got password=%s", user.Password, password)
			log.Printf("Error: %v", err)
			SetError(c, "INVALID_USERNAME_PASSWORD", http.StatusUnauthorized)
			RespondWithError(c, http.StatusUnauthorized, "INVALID_USERNAME_PASSWORD", "Invalid username or password")
			c.Abort()
			return
		}

		// Аутентификация прошла успешно, сохраняем user_id в контексте
		log.Printf("Authentication successful for user: %s", username)
		c.Set("user_id", user.ID) // Добавляем user_id в контекст
		c.Next()
	}
}
