package handlers

import (
	"errors"
	"hetzner-api-emulator/middlewares"
	"hetzner-api-emulator/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UpdateServerName(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userId, err := middlewares.GetUserIDFromContext(c)
        if err != nil || userId == 0 {
            middlewares.RespondWithError(c, http.StatusBadRequest, "USER_ID_MISSING", "User ID is missing")
            return
        }

        // Получаем server_number из параметров
        serverNumberStr := c.Param("server-number")
        serverNumber, err := strconv.Atoi(serverNumberStr)
        if err != nil {
            middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_SERVER_NUMBER", "Invalid server number format")
            return
        }

        // Получаем новое имя сервера
        serverName := c.DefaultPostForm("server_name", "")
        if serverName == "" {
            middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_INPUT", "Server name is required")
            return
        }

        // Обновляем имя сервера
        var server models.Server
        if err := db.Where("user_id = ? AND server_number = ?", userId, serverNumber).First(&server).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                middlewares.RespondWithError(c, http.StatusNotFound, "SERVER_NOT_FOUND", "Server not found")
                return
            }
            middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
            return
        }

        // Обновляем имя сервера
        server.ServerName = serverName
        if err := db.Save(&server).Error; err != nil {
            middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update server name")
            return
        }

        // Формируем ответ
        var serverResponse models.ServerResponse
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

		// Загружаем IP и Subnet
		var ips []models.IP
		if err := db.Where("server_id = ?", server.ID).Find(&ips).Error; err != nil {
			middlewares.RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to load IPs")
			return
		}

		for _, ip := range ips {
			serverResponse.Server.IP = append(serverResponse.Server.IP, ip.IPAddress)
			serverResponse.Server.Subnet = append(serverResponse.Server.Subnet, models.Subnet{
				IP:   ip.IPAddress,
				Mask: ip.Mask,
			})
		}

        // Добавляем дополнительные параметры
        serverResponse.Server.Reset = server.Reset
        serverResponse.Server.Rescue = server.Rescue
        serverResponse.Server.Vnc = server.Vnc
        serverResponse.Server.Windows = server.Windows
        serverResponse.Server.Plesk = server.Plesk
        serverResponse.Server.Cpanel = server.Cpanel
        serverResponse.Server.Wol = server.Wol
        serverResponse.Server.HotSwap = server.HotSwap
        serverResponse.Server.LinkedStoragebox = server.LinkedStoragebox


        c.JSON(http.StatusOK, serverResponse)
    }
}

