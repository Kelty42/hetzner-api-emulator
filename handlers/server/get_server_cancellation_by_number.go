package handlers

import (
	"errors"
	"fmt"
	"hetzner-api-emulator/middlewares"
	"hetzner-api-emulator/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetServerCancellation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := middlewares.GetUserIDFromContext(c)
		if err != nil || userId == 0 {
			middlewares.RespondWithError(c, http.StatusBadRequest, "USER_ID_MISSING", "User ID is missing")
			return
		}

		serverNumberStr := c.Param("server-number")
		serverNumber, err := strconv.Atoi(serverNumberStr)
		if err != nil {
			middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_SERVER_NUMBER", "Invalid server number format")
			return
		}

		var server models.Server
		if err := db.Where("user_id = ? AND server_number = ?", userId, serverNumber).First(&server).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", fmt.Sprintf("Server with number %d not found", serverNumber))
				return
			}
			middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		// Расчёт даты отмены
		earliestCancellationDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

		// Определение cancellation_reason
		var cancellationReason interface{}
		if server.Cancelled {
			cancellationReason = server.CancellationReason // Причина из базы данных
		} else {
			cancellationReason = models.GetAllCancellationReasons() // Массив причин
		}

		// Формируем ответ
		response := gin.H{
			"cancellation": gin.H{
				"server_ip":                server.ServerIP,
				"server_ipv6_net":          server.ServerIPv6Net,
				"server_number":            server.ServerNumber,
				"server_name":              server.ServerName,
				"earliest_cancellation_date": earliestCancellationDate,
				"cancelled":                server.Cancelled,
				"reservation_possible":     server.ReservationPossible,
				"reserved":                 server.Reserved,
				"cancellation_date":        server.CancellationDate,
				"cancellation_reason":      cancellationReason,
			},
		}

		c.JSON(http.StatusOK, response)
	}
}
