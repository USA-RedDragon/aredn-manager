package apimodels

type WebsocketTunnelStats struct {
	ID               uint    `json:"id"`
	RXBytesPerSecond uint64  `json:"rx_bytes_per_sec"`
	TXBytesPerSecond uint64  `json:"tx_bytes_per_sec"`
	RXBytes          uint64  `json:"rx_bytes"`
	TXBytes          uint64  `json:"tx_bytes"`
	TotalRXMB        float64 `json:"total_rx_mb"`
	TotalTXMB        float64 `json:"total_tx_mb"`
}

type WebsocketTunnelConnect struct {
	ID uint `json:"id"`
}

type WebsocketTunnelDisconnect struct {
	ID uint `json:"id"`
}
