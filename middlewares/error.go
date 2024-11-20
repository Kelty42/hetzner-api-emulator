package middlewares

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware для форматирования ошибок
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверяем, возникла ли ошибка
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Получаем значение кода ошибки и статуса из контекста
			errorCode, exists := c.Get("errorCode")
			if !exists {
				errorCode = "UNKNOWN_ERROR"
			}

			errorStatus, exists := c.Get("errorStatus")
			if !exists {
				errorStatus = http.StatusInternalServerError
			}

			// Форматируем ответ в нужном виде
			RespondWithError(c, errorStatus.(int), errorCode.(string), err.Error())
		}
	}
}

// RespondWithError отправляет JSON-ответ с информацией об ошибке
func RespondWithError(c *gin.Context, status int, code string, message string) {
	// Устанавливаем статус ответа
	if status == 0 {
		status = http.StatusInternalServerError
	}

	// Формируем ответ в нужном формате
	c.JSON(status, gin.H{
		"error": gin.H{
			"status":  status,
			"code":    code,
			"message": message,
		},
	})
	c.Abort()
}

// SetError добавляет код ошибки и статус в контекст
func SetError(c *gin.Context, code string, status int) {
	c.Set("errorCode", code)
	c.Set("errorStatus", status)
}
