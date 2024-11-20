package handlers

import (
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes регистрирует маршруты для работы с пользователями
func RegisterUserRoutes(router *gin.RouterGroup) {
	router.GET("/users", GetUsersHandler)        // Получение списка пользователей
	router.GET("/users/:id", GetUserHandler)     // Получение информации о конкретном пользователе
	router.POST("/users", CreateUserHandler)     // Создание нового пользователя
	router.PUT("/users/:id", UpdateUserHandler)  // Обновление пользователя
	router.DELETE("/users/:id", DeleteUserHandler) // Удаление пользователя
}

func GetUsersHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "List of users"})
}

func GetUserHandler(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "Details of user " + id})
}

func CreateUserHandler(c *gin.Context) {
	c.JSON(201, gin.H{"message": "User created"})
}

func UpdateUserHandler(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "User " + id + " updated"})
}

func DeleteUserHandler(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "User " + id + " deleted"})
}
