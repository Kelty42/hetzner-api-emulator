package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"hetzner-api-emulator/middlewares"
)

// Обработчик для отмены отмены сервера
func DeleteServerCancellation(db *sql.DB, dbType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverNumber := c.Param("server-number")

		// Формируем запрос в зависимости от типа базы данных
		var query string
		switch dbType {
		case "postgres":
			query = `SELECT cancelled, reserved FROM servers WHERE server_number = $1`
		case "mysql":
			query = `SELECT cancelled, reserved FROM servers WHERE server_number = ?`
		default:
			middlewares.SetError(c, "INVALID_DB_TYPE", http.StatusInternalServerError)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "INVALID_DB_TYPE", "Invalid database type")
			return
		}

		// Проверяем, существует ли сервер
		var cancelled, reserved bool
		err := db.QueryRow(query, serverNumber).Scan(&cancelled, &reserved)
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

		// Если сервер уже не отменён, возвращаем ошибку конфликта
		if !cancelled {
			middlewares.SetError(c, "CONFLICT", http.StatusConflict)
			middlewares.RespondWithError(c, http.StatusConflict, "CONFLICT", "The cancellation cannot be revoked")
			return
		}

		// Формируем запрос на отмену отмены, также сбрасываем флаг reserved, если он true
		var updateQuery string
		switch dbType {
		case "postgres":
			updateQuery = `UPDATE servers SET cancelled = FALSE, cancellation_date = NULL, cancellation_reason = NULL, reserved = FALSE WHERE server_number = $1`
		case "mysql":
			updateQuery = `UPDATE servers SET cancelled = FALSE, cancellation_date = NULL, cancellation_reason = NULL, reserved = FALSE WHERE server_number = ?`
		}

		// Выполняем запрос на обновление данных
		_, err = db.Exec(updateQuery, serverNumber)
		if err != nil {
			log.Printf("Error updating server cancellation: %v", err)
			middlewares.SetError(c, "INTERNAL_ERROR", http.StatusInternalServerError)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Cancellation revocation failed due to an internal error")
			return
		}

		// Возвращаем успешный ответ без тела
		c.Status(http.StatusNoContent)
	}
}
