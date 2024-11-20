package models

// CancellationReason структура для хранения допустимых причин отмены
type CancellationReason struct {
	Reasons []string `json:"reasons"`
}

// GetAllCancellationReasons возвращает список всех допустимых причин отмены
func GetAllCancellationReasons() []string {
	return []string{
		"Upgrade to a new server",
		"Dissatisfied with the hardware",
		"Dissatisfied with the support",
		"Dissatisfied with the network",
		"Dissatisfied with the IP/subnet allocation",
		"Dissatisfied with the Robot webinterface",
		"Dissatisfied with the official Terms and Conditions",
		"Server no longer necessary due to project ending",
		"Server too expensive",
	}
}
