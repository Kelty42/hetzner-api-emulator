package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"hetzner-api-emulator/middlewares"
)

// Server структура для описания сервера.
type Server struct {
	ServerIP       string        `json:"server_ip"`
	ServerIPv6Net  string        `json:"server_ipv6_net"`
	ServerNumber   int           `json:"server_number"`  // Изменено с string на int
	ServerName     string        `json:"server_name"`
	Product        string        `json:"product"`
	DC             string        `json:"dc"`
	Traffic        string        `json:"traffic"`
	Status         string        `json:"status"`
	Cancelled      bool          `json:"cancelled"`
	PaidUntil      string        `json:"paid_until"`
	IP             []string      `json:"ip"`
	Subnet         []map[string]string `json:"subnet"`
}

// ParseIP преобразует строку IP в массив строк
func ParseIP(ipStr string) []string {
	ipStr = strings.Trim(ipStr, "{}")
	if ipStr == "" {
		return nil
	}
	return strings.Split(ipStr, ",")
}

// ParseSubnet преобразует строку в формат подсети
func ParseSubnet(subnetStr string) []map[string]string {
	var subnet []map[string]string
	if subnetStr == "null" || subnetStr == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(subnetStr), &subnet); err != nil {
		log.Printf("Error parsing subnet: %v", err)
		return nil
	}
	return subnet
}

// GetServers возвращает список серверов для пользователя.
func GetServers(db *sql.DB, dbType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из контекста
		userId, err := middlewares.GetUserIDFromContext(c)
		if err != nil || userId == 0 {
			middlewares.RespondWithError(c, http.StatusBadRequest, "USER_ID_MISSING", "User ID is missing")
			return
		}

		// SQL запрос для получения серверов
		var rows *sql.Rows
		var query string

		if dbType == "postgres" {
			query = `SELECT server_ip, server_ipv6_net, server_number, server_name, product, dc, traffic, status, cancelled, paid_until, ip, subnet 
					 FROM servers WHERE user_id = $1`
			rows, err = db.Query(query, userId)
		} else {
			query = `SELECT server_ip, server_ipv6_net, server_number, server_name, product, dc, traffic, status, cancelled, paid_until, ip, subnet 
					 FROM servers WHERE user_id = ?`
			rows, err = db.Query(query, userId)
		}

		// log.Printf("Executing query: %s with user_id: %d", query, userId)

		if err != nil {
			log.Printf("Error executing query: %v", err)
			middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}
		defer rows.Close()

		// Обработка результатов
		var servers []map[string]interface{}
		for rows.Next() {
			var server Server
			var paidUntil time.Time
			var subnetStr sql.NullString  // Используем sql.NullString для поля subnet
			var ipBytes []byte
			var product, dc, traffic, status sql.NullString
			var cancelled sql.NullBool

			if err := rows.Scan(&server.ServerIP, &server.ServerIPv6Net, &server.ServerNumber, &server.ServerName, 
				&product, &dc, &traffic, &status, &cancelled, &paidUntil, 
				&ipBytes, &subnetStr); err != nil {
				log.Printf("Error scanning row: %v", err)
				middlewares.RespondWithError(c, http.StatusInternalServerError, "SERVER_SCAN_ERROR", "Error scanning server data")
				return
			}

			// Преобразуем sql.NullString в обычные строки
			if product.Valid {
				server.Product = product.String
			}
			if dc.Valid {
				server.DC = dc.String
			}
			if traffic.Valid {
				server.Traffic = traffic.String
			}
			if status.Valid {
				server.Status = status.String
			}

			// Преобразуем sql.NullBool в bool
			if cancelled.Valid {
				server.Cancelled = cancelled.Bool
			}

			// Обрабатываем поле PaidUntil
			if paidUntil.IsZero() {
				server.PaidUntil = ""  // Если дата не задана, ставим пустую строку
			} else {
				server.PaidUntil = paidUntil.Format("2006-01-02")
			}

			// Преобразуем IP
			server.IP = ParseIP(string(ipBytes))

			// Обрабатываем поле Subnet
			if subnetStr.Valid {
				server.Subnet = ParseSubnet(subnetStr.String)
			}

			servers = append(servers, map[string]interface{}{"server": server})
		}

		if len(servers) == 0 {
			middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", "No servers found")
			return
		}

		c.JSON(http.StatusOK, servers)
	}
}
