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

type ServerResponse struct {
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
		Subnet        []Subnet `json:"subnet"`
		Reset         bool     `json:"reset"`
		Rescue        bool     `json:"rescue"`
		Vnc           bool     `json:"vnc"`
		Windows       bool     `json:"windows"`
		Plesk         bool     `json:"plesk"`
		Cpanel        bool     `json:"cpanel"`
		Wol           bool     `json:"wol"`
		HotSwap       bool     `json:"hot_swap"`
		LinkedStoragebox int    `json:"linked_storagebox"`
	} `json:"server"`
}

// Subnet структура для представления подсети
type Subnet struct {
	IP   string `json:"ip"`
	Mask string `json:"mask"`
}