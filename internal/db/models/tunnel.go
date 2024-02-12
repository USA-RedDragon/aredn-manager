package models

import (
	"fmt"
	"net"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"gorm.io/gorm"
)

type Tunnel struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Hostname           string         `json:"hostname" binding:"required"`
	IP                 string         `json:"ip" binding:"required"`
	Password           string         `json:"-" binding:"required"`
	Active             bool           `json:"active"`
	Client             bool           `json:"client"`
	TunnelInterface    string         `json:"-"`
	RXBytes            uint64         `json:"rx_bytes"`
	TXBytes            uint64         `json:"tx_bytes"`
	TotalRXMB          float64        `json:"total_rx_mb"`
	TotalTXMB          float64        `json:"total_tx_mb"`
	RXBytesPerSec      uint64         `json:"rx_bytes_per_sec"`
	TXBytesPerSec      uint64         `json:"tx_bytes_per_sec"`
	Wireguard          bool           `json:"wireguard" gorm:"default:false"`
	WireguardServerKey string         `json:"wireguard_server_key"`
	WireguardPort      uint16         `json:"wireguard_port"`
	ConnectionTime     time.Time      `json:"connection_time"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"-"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}

func TunnelIDExists(db *gorm.DB, id uint) (bool, error) {
	var count int64
	err := db.Model(&Tunnel{}).Where("ID = ?", id).Limit(1).Count(&count).Error
	return count > 0, err
}

func FindTunnelByID(db *gorm.DB, id uint) (Tunnel, error) {
	var tunnel Tunnel
	err := db.First(&tunnel, id).Error
	return tunnel, err
}

func FindTunnelByInterface(db *gorm.DB, iface string) (Tunnel, error) {
	var tunnel Tunnel
	err := db.Where("tunnel_interface = ?", iface).First(&tunnel).Error
	return tunnel, err
}

func FindTunnelByIP(db *gorm.DB, ip net.IP) (Tunnel, error) {
	var tunnel Tunnel
	err := db.Where("ip = ?", ip.String()).First(&tunnel).Error
	return tunnel, err
}

func ListVtunTunnels(db *gorm.DB) ([]Tunnel, error) {
	var tunnels []Tunnel
	err := db.Not("wireguard = ?", true).Order("id asc").Find(&tunnels).Error
	return tunnels, err
}

func ListAllTunnels(db *gorm.DB) ([]Tunnel, error) {
	var tunnels []Tunnel
	err := db.Order("id asc").Find(&tunnels).Error
	return tunnels, err
}

func ListWireguardTunnels(db *gorm.DB) ([]Tunnel, error) {
	var tunnels []Tunnel
	err := db.Where("wireguard = ?", true).Order("id asc").Find(&tunnels).Error
	return tunnels, err
}

func ListClientTunnels(db *gorm.DB) ([]Tunnel, error) {
	var tunnels []Tunnel
	err := db.Where("client = ?", true).Order("id asc").Find(&tunnels).Error
	return tunnels, err
}

func ListServerTunnels(db *gorm.DB) ([]Tunnel, error) {
	var tunnels []Tunnel
	err := db.Where("client = ?", false).Or("client IS NULL").Order("id asc").Find(&tunnels).Error
	return tunnels, err
}

func CountVtunTunnels(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Tunnel{}).Not("wireguard = ?", true).Count(&count).Error
	return int(count), err
}

func CountWireguardTunnels(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Tunnel{}).Where("wireguard = ?", true).Count(&count).Error
	return int(count), err
}

func CountAllActiveTunnels(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Tunnel{}).Where("active = ?", true).Count(&count).Error
	return int(count), err
}

func CountVTunActiveTunnels(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Tunnel{}).Where("active = ?", true).Not("wireguard = ?", true).Count(&count).Error
	return int(count), err
}

func CountWireguardActiveTunnels(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Tunnel{}).Where("active = ?", true).Where("wireguard = ?", true).Count(&count).Error
	return int(count), err
}

func DeleteTunnel(db *gorm.DB, id uint) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		tx.Unscoped().Delete(&Tunnel{ID: id})
		return nil
	})
	if err != nil {
		fmt.Printf("Error deleting tunnel: %v\n", err)
		return err
	}
	return nil
}

func ClearActiveFromAllTunnels(db *gorm.DB) error {
	return db.Model(&Tunnel{}).Where("active = ?", true).Update("active", false).Error
}

func GetNextVTunIP(db *gorm.DB, config *config.Config) (string, error) {
	// Each tunnel is added with an ip starting from config.VTUNStartingAddress and incrementing by 4 for each tunnel
	// We need to find the next available ip.
	var tunnels []Tunnel
	err := db.Not("wireguard = ?", true).Find(&tunnels).Error
	if err != nil {
		return "", err
	}
	// We need to find the next available ip.
	// We can do this by finding the highest ip, and adding 4 to it.
	var highestIP = net.ParseIP(config.VTUNStartingAddress).To4() // Use 12 so the +4 later starts at 16
	for _, tunnel := range tunnels {
		ip := net.ParseIP(tunnel.IP)
		ip = ip.To4()
		if ip[2] > highestIP[2] {
			highestIP = ip
		} else if ip[2] == highestIP[2] {
			if ip[3] > highestIP[3] {
				highestIP = ip
			}
		}
	}
	// If the highest ip is 252, we need to start at highestIP[2]++ and set highestIP[3] to 0.
	if highestIP[3] == 252 {
		highestIP[2]++
		if highestIP[2] >= 254 {
			return "", fmt.Errorf("no more IPs available")
		}
		highestIP[3] = 0
	} else {
		highestIP[3] += 4
	}

	return highestIP.String(), nil
}

func GetNextWireguardIP(db *gorm.DB, config *config.Config) (string, error) {
	// Each tunnel is added with an ip starting from config.WireguardStartingAddress and incrementing by 4 for each tunnel
	// We need to find the next available ip.
	var tunnels []Tunnel
	err := db.Where("wireguard = ?", true).Find(&tunnels).Error
	if err != nil {
		return "", err
	}
	// We need to find the next available ip.
	// We can do this by finding the highest ip, and adding 4 to it.
	var highestIP = net.ParseIP(config.WireguardStartingAddress).To4() // Use 12 so the +4 later starts at 16
	for _, tunnel := range tunnels {
		ip := net.ParseIP(tunnel.IP)
		ip = ip.To4()
		if ip[2] > highestIP[2] {
			highestIP = ip
		} else if ip[2] == highestIP[2] {
			if ip[3] > highestIP[3] {
				highestIP = ip
			}
		}
	}
	// If the highest ip is 252, we need to start at highestIP[2]++ and set highestIP[3] to 0.
	if highestIP[3] == 252 {
		highestIP[2]++
		if highestIP[2] >= 254 {
			return "", fmt.Errorf("no more IPs available")
		}
		highestIP[3] = 0
	} else {
		highestIP[3] += 4
	}

	return highestIP.String(), nil
}

func GetNextWireguardPort(db *gorm.DB, config *config.Config) (uint16, error) {
	// Each tunnel is added with a port starting from 51820 and incrementing by 1 for each tunnel
	// We need to find the next available port.
	var tunnels []Tunnel
	err := db.Where("wireguard = ?", true).Find(&tunnels).Error
	if err != nil {
		return 0, err
	}
	// We need to find the next available port.
	// We can do this by finding the highest port, and adding 1 to it.
	var highestPort uint16 = config.WireguardStartingPort - 1
	for _, tunnel := range tunnels {
		if tunnel.WireguardPort > highestPort {
			highestPort = tunnel.WireguardPort
		}
	}
	// If the highest port is 65535, we need to return an error.
	if highestPort == 65535 {
		return 0, fmt.Errorf("no more ports available")
	}

	return highestPort + 1, nil
}
