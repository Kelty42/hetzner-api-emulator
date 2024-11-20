package middlewares

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// GetUserIDFromContext извлекает user_id из контекста запроса
func GetUserIDFromContext(c *gin.Context) (int, error) {
	// Попробуем извлечь user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("User ID not found in context")
		return 0, http.ErrNoLocation // Возвращаем ошибку, если user_id не найден
	}

	// Преобразуем в целое число и возвращаем
	if id, ok := userID.(int); ok {
		return id, nil
	}
	return 0, http.ErrNoLocation
}
