package models

import "database/sql"

// ServerDataCancellation - структура для хранения информации о сервере и данных отмены.
type ServerDataCancellation struct {
	ServerIP             string         `json:"server_ip"`
	ServerIPv6Net        string         `json:"server_ipv6_net"`
	ServerNumber         int            `json:"server_number"`
	ServerName           string         `json:"server_name"`
	Cancelled            bool           `json:"cancelled"`
	ReservationPossible  bool           `json:"reservation_possible"`
	Reserved             bool           `json:"reserved"`
	CancellationDate     sql.NullString `json:"cancellation_date"`
	CancellationReason   sql.NullString `json:"cancellation_reason"` // Изменено на sql.NullString
}
