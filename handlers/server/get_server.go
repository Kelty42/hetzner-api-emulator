package handlers

import (
	"hetzner-api-emulator/middlewares"
	"hetzner-api-emulator/models"
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


// ServerResponse структура для отдачи данных в нужном формате
type ServerResponseShort struct {
	Server struct {
		ServerIP      string   `json:"server_ip"`
		ServerIPv6Net string   `json:"server_ipv6_net"`
		ServerNumber  int      `json:"server_number"`
		ServerName    string   `json:"server_name"`
		Product       string   `json:"product"`
		DC            string   `json:"dc"`
		Traffic       string   `json:"traffic"`
		Status        string   `json:"status"`
		Cancelled     bool     `json:"cancelled"`
		PaidUntil     string   `json:"paid_until"`
		IP            []string `json:"ip"`
		Subnet        []models.Subnet `json:"subnet"`
	} `json:"server"`
}

func GetServers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userId, err := middlewares.GetUserIDFromContext(c)
        if err != nil || userId == 0 {
            middlewares.RespondWithError(c, http.StatusBadRequest, "USER_ID_MISSING", "User ID is missing")
            return
        }

        var servers []models.Server
        if err := db.Where("user_id = ?", userId).
            Preload("IPs").
            Find(&servers).Error; err != nil {
            middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve servers")
            return
        }

        // Если у пользователя нет серверов
        if len(servers) == 0 {
            middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", "No server found")
            return
        }

        var serverResponses []ServerResponseShort
        for _, server := range servers {
            var serverResponse ServerResponseShort
            // Заполнение информации о сервере
            serverResponse.Server.ServerNumber = server.ServerNumber
            serverResponse.Server.ServerName = server.ServerName
            serverResponse.Server.Product = server.Product
            serverResponse.Server.DC = server.DC
            serverResponse.Server.Traffic = server.Traffic
            serverResponse.Server.Status = server.Status
            serverResponse.Server.Cancelled = server.Cancelled
            if server.PaidUntil != nil {
                serverResponse.Server.PaidUntil = server.PaidUntil.Format("2006-01-02")
            }
            serverResponse.Server.ServerIP = server.ServerIP
            serverResponse.Server.ServerIPv6Net = server.ServerIPv6Net
        
            // Получаем IP-адреса для сервера из связанной таблицы IPs
            var ipAddresses []string
            for _, ip := range server.IPs {
                ipAddresses = append(ipAddresses, ip.IPAddress)
            }
            serverResponse.Server.IP = ipAddresses
        
            // Получаем подсети для сервера
            var subnets []models.Subnet
            for _, ip := range server.IPs {
                subnets = append(subnets, models.Subnet{
                    IP:   ip.IPAddress,
                    Mask: ip.Mask,
                })
            }
            if len(subnets) > 0 {
                serverResponse.Server.Subnet = subnets
            } else {
                serverResponse.Server.Subnet = nil
            }
        
            serverResponses = append(serverResponses, serverResponse)
        }

        c.JSON(http.StatusOK, serverResponses)
    }
}
