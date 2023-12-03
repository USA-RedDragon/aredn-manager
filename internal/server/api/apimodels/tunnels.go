package apimodels

import (
	"regexp"
	"time"
)

const minHostnameLength = 3
const maxHostnameLength = 63

type CreateTunnel struct {
	Hostname string `json:"hostname" binding:"required"`
	Password string `json:"password" binding:"required"`
	IP       string `json:"ip"`
	Client   bool   `json:"client"`
}

func (r *CreateTunnel) IsValidHostname() (bool, string) {
	if len(r.Hostname) < minHostnameLength {
		return false, "Hostname must be at least 3 characters"
	}
	if len(r.Hostname) > maxHostnameLength {
		return false, "Hostname must be less than 64 characters"
	}
	if !regexp.MustCompile(`^[A-Za-z0-9\-]+$`).MatchString(r.Hostname) {
		return false, "Hostname must be alphanumeric or -"
	}
	return true, ""
}

type TunnelWithPass struct {
	ID             uint      `json:"id"`
	Client         bool      `json:"client"`
	Hostname       string    `json:"hostname"`
	IP             string    `json:"ip"`
	Password       string    `json:"password"`
	Active         bool      `json:"active"`
	ConnectionTime time.Time `json:"connection_time"`
	CreatedAt      time.Time `json:"created_at"`
}

type EditTunnel struct {
	ID       uint   `json:"id" binding:"required"`
	Hostname string `json:"hostname" binding:"required"`
	Password string `json:"password" binding:"required"`
	IP       string `json:"ip" binding:"required"`
}
