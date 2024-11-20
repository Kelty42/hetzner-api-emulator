package config

import (
	"os"
)

type Config struct {
	DBType       string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	Host         string
	Port         string
}

// LoadConfig загружает конфигурацию приложения из переменных окружения
func LoadConfig() *Config {
	return &Config{
		DBType:       getEnv("DB_TYPE", "postgres"),     // Тип базы данных: "postgres" или "mysql"
		DBHost:       getEnv("DB_HOST", "localhost"),   // Хост базы данных
		DBPort:       getEnv("DB_PORT", "5432"),        // Порт базы данных
		DBUser:       getEnv("DB_USER", "myuser"),      // Имя пользователя базы данных
		DBPassword:   getEnv("DB_PASSWORD", "mypassword"), // Пароль пользователя базы данных
		DBName:       getEnv("DB_NAME", "mydatabase"),  // Название базы данных
		Host:         getEnv("HOST", "0.0.0.0"),        // Хост приложения
		Port:         getEnv("PORT", "8080"),           // Порт приложения
	}
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
