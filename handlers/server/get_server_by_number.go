package handlers

import (
	"fmt"
	"hetzner-api-emulator/middlewares"
	"hetzner-api-emulator/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetServerByNumber(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := middlewares.GetUserIDFromContext(c)
		if err != nil || userId == 0 {
			middlewares.RespondWithError(c, http.StatusBadRequest, "USER_ID_MISSING", "User ID is missing")
			return
		}

		serverNumber := c.Param("server-number")

		var server models.Server
		if err := db.Where("server_number = ?", serverNumber).First(&server).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", fmt.Sprintf("Server with id %s not found", serverNumber))
			} else {
				middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			}
			return
		}

		// Заполняем базовые данные
		response := models.ServerResponse{}

		// Заполняем базовые данные
		response.Server.ServerIP = server.ServerIP
		response.Server.ServerNumber = server.ServerNumber
		response.Server.ServerName = server.ServerName
		response.Server.Product = server.Product
		response.Server.DC = server.DC
		response.Server.Traffic = server.Traffic
		response.Server.Status = server.Status
		response.Server.Cancelled = server.Cancelled
		response.Server.PaidUntil = server.PaidUntil.Format("2006-01-02")
		response.Server.ServerIPv6Net = server.ServerIPv6Net

		// Прямо добавляем экстра параметры
		response.Server.Reset = server.Reset
		response.Server.Rescue = server.Rescue
		response.Server.Vnc = server.Vnc
		response.Server.Windows = server.Windows
		response.Server.Plesk = server.Plesk
		response.Server.Cpanel = server.Cpanel
		response.Server.Wol = server.Wol
		response.Server.HotSwap = server.HotSwap
		response.Server.LinkedStoragebox = server.LinkedStoragebox

		// Заполняем IP и Subnet
		var ips []models.IP
		if err := db.Where("server_id = ?", server.ID).Find(&ips).Error; err != nil {
			middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to load IPs")
			return
		}

		for _, ip := range ips {
			response.Server.IP = append(response.Server.IP, ip.IPAddress)
			response.Server.Subnet = append(response.Server.Subnet, models.Subnet{
				IP:   ip.IPAddress,
				Mask: ip.Mask,
			})
		}

		// Отправляем финальный JSON
		c.JSON(http.StatusOK, response)
	}
}
