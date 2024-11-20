package main

import (
	"log"
	"net/http"

	"hetzner-api-emulator/config"
	"hetzner-api-emulator/database"
	"hetzner-api-emulator/middlewares"
	"hetzner-api-emulator/routes" // Правильный импорт пакета routes
	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Инициализируем подключение к базе данных
	db, status, code, message := database.InitDB(cfg)
	if status != http.StatusOK {
		// Логируем ошибку
		log.Printf("Error: %s - %s", code, message)

		// Создаем роутер Gin
		router := gin.Default()

		// Обработчик ошибок
		router.Use(middlewares.ErrorHandler())

		// Отправляем ошибку
		router.NoRoute(func(c *gin.Context) {
			middlewares.RespondWithError(c, status, code, message)
		})

		// Прекращаем выполнение, если ошибка подключения
		router.Run(cfg.Host + ":" + cfg.Port)
		return
	}

	// Создаем роутер Gin
	router := gin.Default()

	// Обработчик для несуществующих маршрутов
	router.NoRoute(func(c *gin.Context) {
		middlewares.RespondWithError(c, http.StatusNotFound, "ROUTE_NOT_FOUND", "Route not found")
	})

	// Подключаем обработчик ошибок
	router.Use(middlewares.ErrorHandler())

	// Добавляем middleware авторизации с указанием типа базы данных
	authorized := router.Group("/", middlewares.DBAuthMiddleware(db, cfg.DBType))

	// Регистрируем все маршруты через RegisterAllRoutes
	routes.RegisterAllRoutes(authorized, db, cfg.DBType) // Используем правильный вызов из пакета routes

	// Запускаем сервер
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Starting server at %s...", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
