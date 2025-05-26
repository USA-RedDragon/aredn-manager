package apimodels

import "time"

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
	ID     uint `json:"id"`
	Client bool `json:"client"`
	ConnectionTime time.Time `json:"connection_time"`
}

type WebsocketTunnelDisconnect struct {
	ID     uint `json:"id"`
	Client bool `json:"client"`
}

type WebsocketTotalBandwidth struct {
	RX uint64 `json:"RX"`
	TX uint64 `json:"TX"`
}

type WebsocketTotalTraffic struct {
	RX float64 `json:"RX"`
	TX float64 `json:"TX"`
}