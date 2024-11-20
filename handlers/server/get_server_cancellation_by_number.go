package handlers

import (
	"encoding/json" // Добавляем для работы с JSON
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"hetzner-api-emulator/middlewares"
	"hetzner-api-emulator/models" // Импортируем модель
)

// Обработчик для получения данных о сервере и отмене
func GetServerCancellation(db *sql.DB, dbType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverNumber := c.Param("server-number")

		// Формируем запрос в зависимости от типа базы данных
		var query string
		switch dbType {
		case "postgres":
			query = `SELECT server_ip, server_ipv6_net, server_number, server_name, cancelled, reservation_possible, reserved, cancellation_date, cancellation_reason
					  FROM servers WHERE server_number = $1`
		case "mysql":
			query = `SELECT server_ip, server_ipv6_net, server_number, server_name, cancelled, reservation_possible, reserved, cancellation_date, cancellation_reason
					  FROM servers WHERE server_number = ?`
		default:
			middlewares.SetError(c, "INVALID_DB_TYPE", http.StatusInternalServerError)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "INVALID_DB_TYPE", "Invalid database type")
			return
		}

		// Выполняем запрос к базе данных
		var serverData models.ServerDataCancellation
		err := db.QueryRow(query, serverNumber).Scan(
			&serverData.ServerIP,
			&serverData.ServerIPv6Net,
			&serverData.ServerNumber,
			&serverData.ServerName,
			&serverData.Cancelled,
			&serverData.ReservationPossible,
			&serverData.Reserved,
			&serverData.CancellationDate,
			&serverData.CancellationReason,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				middlewares.SetError(c, "SERVER_NOT_FOUND", http.StatusNotFound)
				middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", "Server with id "+serverNumber+" not found")
				return
			}
			log.Printf("Error querying database: %v", err)
			middlewares.SetError(c, "DB_QUERY_ERROR", http.StatusInternalServerError)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "DB_QUERY_ERROR", "Error retrieving server data")
			return
		}

		// Статичная дата отмены через 7 дней
		sevenDaysFromNow := time.Now().Add(7 * time.Hour * 24).Format("2006-01-02")

		// Если в названии продукта есть "DS", разрешаем резервирование
		if containsDS(serverData.ServerName) {
			serverData.ReservationPossible = true
		} else {
			serverData.ReservationPossible = false
		}

		var cancellationReason interface{}
		if serverData.Cancelled {
			if serverData.CancellationReason.Valid {
				// Получаем строку причины
				reasonStr := serverData.CancellationReason.String
				
				// Пробуем распарсить как JSON только если это действительно нужно
				var reasons []string
				if reasonStr[0] == '[' {
					// Если строка выглядит как JSON-массив, распарсиваем её
					err := json.Unmarshal([]byte(reasonStr), &reasons)
					if err != nil {
						log.Printf("Error unmarshalling cancellation reason: %v", err)
						cancellationReason = reasonStr // В случае ошибки возвращаем как строку
					} else {
						cancellationReason = reasons
					}
				} else {
					// Если это обычная строка, просто передаём её как строку
					cancellationReason = reasonStr
				}
			} else {
				cancellationReason = "" // Если причины нет, отправляем пустую строку
			}
		} else {
			// Если не отменён, передаём все возможные причины
			cancellationReason = models.GetAllCancellationReasons()
		}
		



		// Возвращаем данные о сервере в нужном формате
		c.JSON(http.StatusOK, gin.H{
			"cancellation": gin.H{
				"server_ip":               serverData.ServerIP,
				"server_ipv6_net":         serverData.ServerIPv6Net,
				"server_number":           serverData.ServerNumber,
				"server_name":             serverData.ServerName,
				"earliest_cancellation_date": sevenDaysFromNow, // Статичная дата
				"cancelled":               serverData.Cancelled,
				"reservation_possible":    serverData.ReservationPossible,
				"reserved":                serverData.Reserved,
				"cancellation_date":       serverData.CancellationDate.String, // Используем строку из sql.NullString
				"cancellation_reason":     cancellationReason, // Строка или массив, в зависимости от отмены
			},
		})
	}
}

// Проверка, содержится ли DS в названии продукта
func containsDS(productName string) bool {
	return strings.Contains(productName, "DS")
}
