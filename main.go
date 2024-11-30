package main

import (
	"flag"
	"log"
	"net/http"

	"hetzner-api-emulator/config"
	"hetzner-api-emulator/database"
	"hetzner-api-emulator/models"
	"hetzner-api-emulator/middlewares"
	"hetzner-api-emulator/routes" // Правильный импорт пакета routes
	"github.com/gin-gonic/gin"
)

func main() {
	// Добавляем флаг для миграции
	migrateFlag := flag.Bool("migration", false, "Run migrations")
	flag.Parse()
	cfg := config.LoadConfig()
	// Если флаг миграции установлен, выполняем миграции и выходим
	if *migrateFlag {

		// Инициализируем подключение к базе данных
		database.ConnectDatabase()

		// Получаем соединение с БД
		db := database.GetDB()

		// Запускаем миграции
		models.Migrate(db)

		log.Println("Migrations completed successfully")
		return
	}

	// Инициализируем подключение к базе данных
	database.ConnectDatabase()

	// Создаем роутер Gin
	router := gin.Default()

	// Обработчик для несуществующих маршрутов
	router.NoRoute(func(c *gin.Context) {
		middlewares.RespondWithError(c, http.StatusNotFound, "ROUTE_NOT_FOUND", "Route not found")
	})

	// Подключаем обработчик ошибок
	router.Use(middlewares.ErrorHandler())

	// Добавляем middleware авторизации с указанием типа базы данных
	authorized := router.Group("/", middlewares.DBAuthMiddleware(database.GetDB()))

	// Регистрируем все маршруты через RegisterAllRoutes
	routes.RegisterAllRoutes(authorized, database.GetDB(), cfg.DBType) // Используем правильный вызов из пакета routes

	// Запускаем сервер
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Starting server at %s...", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
