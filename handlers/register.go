package handlers

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// User структура для представления пользователя
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterUserHandler обрабатывает регистрацию нового пользователя
func RegisterUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем базу данных из контекста
		db, exists := c.Get("db")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database not found"})
			return
		}

		// Парсим тело запроса в структуру
		var user User
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

		// Получаем объект базы данных
		gormDB := db.(*gorm.DB)

		// Создаем нового пользователя
		newUser := User{
			Username: user.Username,
			Password: string(hashedPassword),
		}

		// Сохраняем пользователя в базу данных
		if err := gormDB.Create(&newUser).Error; err != nil {
			log.Printf("Database insertion error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
			return
		}

		// Успешный ответ
		log.Printf("User %s registered successfully", user.Username)
		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}
