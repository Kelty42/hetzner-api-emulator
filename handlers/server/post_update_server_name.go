package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"

	"hetzner-api-emulator/middlewares"

	"github.com/gin-gonic/gin"
)

// UpdateServerName обновляет имя сервера
func UpdateServerName(db *sql.DB, dbType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из контекста
		userId, err := middlewares.GetUserIDFromContext(c)
		if err != nil || userId == 0 {
			middlewares.RespondWithError(c, http.StatusBadRequest, "USER_ID_MISSING", "User ID is missing")
			return
		}

		// Получаем server_number из параметров запроса
		serverNumberStr := c.Param("server-number")
		serverNumber, err := strconv.Atoi(serverNumberStr)
		if err != nil {
			middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_SERVER_NUMBER", "Invalid server number format")
			return
		}

		// Получаем новое имя сервера из данных формы
		serverName := c.DefaultPostForm("server_name", "")
		if serverName == "" {
			middlewares.RespondWithError(c, http.StatusBadRequest, "MISSING_SERVER_NAME", "Server name is required")
			return
		}

		// Формируем запрос в зависимости от типа базы данных
		var query string
		switch dbType {
		case "postgres":
			query = `UPDATE servers SET server_name = $1 WHERE user_id = $2 AND server_number = $3`
		case "mysql":
			query = `UPDATE servers SET server_name = ? WHERE user_id = ? AND server_number = ?`
		default:
			middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_DB_TYPE", "Invalid DB type")
			return
		}

		// Выполняем обновление
		_, err = db.Exec(query, serverName, userId, serverNumber)
		if err != nil {
			log.Printf("Error updating server name: %v", err)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Error updating server name")
			return
		}

		// После успешного обновления возвращаем обновленные данные о сервере
		var serverData ServerData
		var paidUntil sql.NullTime
		var subnetStr sql.NullString
		var ipBytes []byte
		var product, dc, traffic, status sql.NullString
		var cancelled sql.NullBool
		var extraParamsJSON []byte

		// Получаем данные о сервере из базы
		var row *sql.Row
		switch dbType {
		case "postgres":
			query = `SELECT server_ip, server_ipv6_net, server_number, server_name, product, dc, traffic, status, cancelled, paid_until, ip, subnet, extra_params 
					 FROM servers WHERE user_id = $1 AND server_number = $2`
			row = db.QueryRow(query, userId, serverNumber)
		case "mysql":
			query = `SELECT server_ip, server_ipv6_net, server_number, server_name, product, dc, traffic, status, cancelled, paid_until, ip, subnet, extra_params 
					 FROM servers WHERE user_id = ? AND server_number = ?`
			row = db.QueryRow(query, userId, serverNumber)
		}

		if err := row.Scan(&serverData.ServerIP, &serverData.ServerIPv6Net, &serverData.ServerNumber, &serverData.ServerName,
			&product, &dc, &traffic, &status, &cancelled, &paidUntil,
			&ipBytes, &subnetStr, &extraParamsJSON); err != nil {
			if err == sql.ErrNoRows {
				middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", "Server not found")
			} else {
				log.Printf("Error scanning row: %v", err)
				middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			}
			return
		}

		// Обрабатываем данные
		if product.Valid {
			serverData.Product = product.String
		}
		if dc.Valid {
			serverData.DC = dc.String
		}
		if traffic.Valid {
			serverData.Traffic = traffic.String
		}
		if status.Valid {
			serverData.Status = status.String
		}
		if cancelled.Valid {
			serverData.Cancelled = cancelled.Bool
		}
		if paidUntil.Valid {
			serverData.PaidUntil = paidUntil.Time.Format("2006-01-02")
		}
		serverData.IP = ParseIP(string(ipBytes))  // Используем уже существующую функцию ParseIP
		if subnetStr.Valid {
			serverData.Subnet = ParseSubnet(subnetStr.String)  // Используем уже существующую функцию ParseSubnet
		}

		// Десериализация дополнительных параметров
		var extraParams map[string]interface{}
		if err := json.Unmarshal(extraParamsJSON, &extraParams); err != nil {
			log.Printf("Error unmarshalling extra_params JSON: %v", err)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "INVALID_EXTRA_PARAMS", "Invalid extra parameters")
			return
		}

		// Формируем ответ с нужным порядком параметров
		response := map[string]interface{}{ "server": gin.H{
			"server_ip":       serverData.ServerIP,
			"server_ipv6_net": serverData.ServerIPv6Net,
			"server_number":   serverData.ServerNumber,
			"server_name":     serverData.ServerName,
			"product":         serverData.Product,
			"dc":              serverData.DC,
			"traffic":         serverData.Traffic,
			"status":          serverData.Status,
			"cancelled":       serverData.Cancelled,
			"paid_until":      serverData.PaidUntil,
			"ip":              serverData.IP,
			"subnet":          serverData.Subnet,
		}}

		// Сортировка ключей в extraParams
		var sortedKeys []string
		for key := range extraParams {
			sortedKeys = append(sortedKeys, key)
		}

		// Сортировка ключей по алфавиту
		sort.Strings(sortedKeys)

		// Добавление параметров в ответ
		for _, key := range sortedKeys {
			response["server"].(gin.H)[key] = extraParams[key]
		}

		// Теперь отправляем отсортированный ответ
		c.JSON(http.StatusOK, response)
	}
}
