package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"hetzner-api-emulator/models"
	"hetzner-api-emulator/middlewares"
)

func PostServerCancellation(db *sql.DB, dbType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverNumber := c.Param("server-number")

		// Проверяем тело запроса
		var request struct {
			CancellationDate   string  `form:"cancellation_date"`
			CancellationReason *string `form:"cancellation_reason"` // используем указатель на строку
			ReserveLocation    string  `form:"reserve_location"`
		}

		if err := c.ShouldBind(&request); err != nil {
			middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
			return
		}

		// Проверяем, является ли причина отмены допустимой
		validReasons := models.GetAllCancellationReasons()
		isValidReason := false
		if request.CancellationReason != nil {
			for _, reason := range validReasons {
				if *request.CancellationReason == reason {
					isValidReason = true
					break
				}
			}
		}

		if !isValidReason && request.CancellationReason != nil {
			middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_CANCELLATION_REASON", "Invalid cancellation reason")
			return
		}

		// Проверяем сервер в базе
		var serverData models.ServerDataCancellation
		query := `SELECT server_ip, server_ipv6_net, server_number, server_name, cancelled, reservation_possible, reserved, cancellation_date, cancellation_reason FROM servers WHERE server_number = $1`
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
				middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", "Server with id "+serverNumber+" not found")
				return
			}
			log.Printf("Error querying database: %v", err)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}

		// Проверяем состояние отмены
		if serverData.Cancelled {
			middlewares.RespondWithError(c, http.StatusConflict, "CONFLICT", "The server is already cancelled")
			return
		}

		// Проверка параметра reserve_location
		if request.ReserveLocation == "true" && !serverData.ReservationPossible {
			middlewares.RespondWithError(c, http.StatusConflict, "SERVER_CANCELLATION_RESERVE_LOCATION_FALSE_ONLY", "It is not possible to reserve the location. Remove parameter reserve_location or set value to 'false'")
			return
		}

		// Устанавливаем дату отмены
		var cancellationDate time.Time
		if request.CancellationDate == "" {
			cancellationDate = time.Now().Add(7 * 24 * time.Hour).Truncate(24 * time.Hour) // Если дата не передана, присваиваем +7 дней
		} else {
			var err error
			cancellationDate, err = time.Parse("2006-01-02", request.CancellationDate)
			if err != nil {
				middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_CANCELLATION_DATE", "Invalid cancellation date format, expected yyyy-MM-dd")
				return
			}

			minCancellationDate := time.Now().Add(96 * time.Hour).Truncate(24 * time.Hour) // Минимальная дата отмены через 4 дня
			if cancellationDate.Before(minCancellationDate) {
				middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_CANCELLATION_DATE", "Cancellation date must be at least 4 days from now")
				return
			}
		}

		// Обновляем данные в базе
		updateQuery := `UPDATE servers SET cancelled = TRUE, cancellation_date = $1, cancellation_reason = $2, reserved = $3 WHERE server_number = $4`
		reserved := strings.ToLower(request.ReserveLocation) == "true"

		// Если причина отмены не передана, передаем NULL в базе
		var cancellationReason interface{}
		if request.CancellationReason != nil {
			// Преобразуем строку в формат, подходящий для базы данных
			cancellationReason = *request.CancellationReason
		} else {
			cancellationReason = nil
		}

		_, err = db.Exec(updateQuery, cancellationDate.Format("2006-01-02"), cancellationReason, reserved, serverNumber)
		if err != nil {
			log.Printf("Error updating database: %v", err)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Cancellation failed due to an internal error")
			return
		}

		// Формируем ответ
		c.JSON(http.StatusOK, gin.H{
			"cancellation": gin.H{
				"server_ip":               serverData.ServerIP,
				"server_ipv6_net":         serverData.ServerIPv6Net,
				"server_number":           serverData.ServerNumber,
				"server_name":             serverData.ServerName,
				"earliest_cancellation_date": cancellationDate.Format("2006-01-02"),
				"cancelled":               true,
				"reserved":                reserved,
				"reservation_possible":    serverData.ReservationPossible,
				"cancellation_date":       cancellationDate.Format("2006-01-02"),
				"cancellation_reason":     request.CancellationReason, // возвращаем как *string
			},
		})
	}
}
