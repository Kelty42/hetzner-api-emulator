package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"github.com/gin-gonic/gin"
)

// RegisterUserHandler обрабатывает регистрацию нового пользователя
func RegisterUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем базу данных и тип базы данных из контекста
		db, exists := c.Get("db")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database not found"})
			return
		}
		dbType, exists := c.Get("dbType")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database type not found"})
			return
		}

		// Структура для получения данных из запроса
		type User struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		var user User
		// Парсим тело запроса в структуру
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Printf("Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
			return
		}

		// Формируем SQL-запрос в зависимости от типа базы данных
		var query string
		switch dbType.(string) {
		case "mysql":
			query = "INSERT INTO users (username, password) VALUES (?, ?)"
		case "postgres":
			query = "INSERT INTO users (username, password) VALUES ($1, $2)"
		default:
			log.Println("Unsupported database type")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unsupported database type"})
			return
		}

		// Выполняем запрос
		_, err = db.(*sql.DB).Exec(query, user.Username, string(hashedPassword))
		if err != nil {
			log.Printf("Database insertion error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
			return
		}

		// Успешный ответ
		log.Printf("User %s registered successfully", user.Username)
		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}
