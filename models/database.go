package models

import (
	"log"
	"os"
	"reflect"
	"time"

	"gorm.io/gorm"
)

type IP struct {
    ID        int       `gorm:"primaryKey;autoIncrement"`
    ServerID  int       `gorm:"not null;index"`
    IPAddress string    `gorm:"type:varchar(15);not null"`
    Mask      string    `gorm:"type:varchar(15);not null"`
}

type Server struct {
    ID                   int       `gorm:"primaryKey;autoIncrement"`
    UserID               int       `gorm:"not null"`
    ServerNumber         int       `gorm:"not null;unique"`
    ServerName           string    `gorm:"type:varchar(255);not null"`
    ServerIP             string    `gorm:"type:varchar(255);"`
    Product              string    `gorm:"type:varchar(255);"`
    ServerIPv6Net        string    `gorm:"type:varchar(255);"`
    DC                   string    `gorm:"type:varchar(255);"`
    Traffic              string    `gorm:"type:varchar(255);"`
    Status               string    `gorm:"type:varchar(255);"`
    Cancelled            bool      `gorm:"default:false"`
    PaidUntil            *time.Time
    IPs                  []IP      `gorm:"foreignKey:ServerID"`
    Reset                bool      `gorm:"column:reset"`
    Rescue               bool      `gorm:"column:rescue"`
    Vnc                  bool      `gorm:"column:vnc"`
    Windows              bool      `gorm:"column:windows"`
    Plesk                bool      `gorm:"column:plesk"`
    Cpanel               bool      `gorm:"column:cpanel"`
    Wol                  bool      `gorm:"column:wol"`
    HotSwap              bool      `gorm:"column:hot_swap"`
    LinkedStoragebox     int       `gorm:"column:linked_storagebox"`
    ReservationPossible   bool       `gorm:"default:false"`
    Reserved              bool       `gorm:"default:false"`
    CancellationDate      *time.Time `gorm:"column:cancellation_date"`
	CancellationReason string `gorm:"type:varchar(255);"`
}

type User struct {
    ID        int       `gorm:"primaryKey;autoIncrement"`
    Username  string    `gorm:"uniqueIndex;type:varchar(255)"`
    Password  string    `gorm:"type:varchar(255)"`
    CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// Migrate функция для миграции моделей в зависимости от типа базы данных
// Migrate функция для миграции моделей в зависимости от типа базы данных
func Migrate(db *gorm.DB) {
	// Получаем тип базы данных из переменной окружения
	dbType := os.Getenv("DB_CONNECTION")
	if dbType == "" {
		log.Fatal("DB_CONNECTION environment variable not set")
	}

	// Получаем список всех моделей
	models := []interface{}{
		&User{},
		&Server{},
		&IP{},
	}

	// Выполняем миграцию для каждой модели
	for _, model := range models {
		if dbType == "mysql" || dbType == "postgres" {
			// Проверяем существование таблицы перед миграцией
			if !db.Migrator().HasTable(model) {
				log.Printf("Table for model %s doesn't exist, creating it", reflect.TypeOf(model).String())
			}

			// Выполняем миграцию
			if err := db.AutoMigrate(model); err != nil {
				log.Fatalf("Migration failed for %s: %v", reflect.TypeOf(model).String(), err)
			}
		} else {
			log.Fatalf("Unsupported DB type: %v", dbType)
		}
	}
}

