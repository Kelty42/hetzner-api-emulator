package routes

import (
	"database/sql"
	serverHandlers "hetzner-api-emulator/handlers/server"
	"hetzner-api-emulator/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterAllRoutes(router *gin.RouterGroup, db *sql.DB, dbType string) {
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("dbType", dbType)
		c.Next()
	})

	RegisterUserRoutes(router)
	RegisterServerRoutes(router.Group("/server"), db, dbType)
}

func RegisterUserRoutes(router *gin.RouterGroup) {
	router.POST("/register", handlers.RegisterUserHandler())
}

func RegisterServerRoutes(serverRouter *gin.RouterGroup, db *sql.DB, dbType string) {
	// Регистрация маршрута для получения списка серверов
	serverRouter.GET("", serverHandlers.GetServers(db, dbType)) // Список серверов
	// Маршрут для получения сервера по номеру
	serverRouter.GET("/:server-number", serverHandlers.GetServerByNumber(db, dbType)) // Сервер по номеру
	// Маршрут для обновления имени сервера
	serverRouter.POST("/:server-number", serverHandlers.UpdateServerName(db, dbType)) // Обновление имени
	// Маршрут для получения информации об отмене сервера
	serverRouter.GET("/:server-number/cancellation", serverHandlers.GetServerCancellation(db, dbType)) // Отмена сервера
	// Маршрут для получения информации об отмене сервера
	serverRouter.POST("/:server-number/cancellation", serverHandlers.PostServerCancellation(db, dbType)) // Отмена сервера
	// Новый маршрут для отмены отмены
	serverRouter.DELETE("/:server-number/cancellation", serverHandlers.DeleteServerCancellation(db, dbType)) // Отмена отмены
}
