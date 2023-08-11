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
}

func (r *CreateTunnel) IsValidHostname() (bool, string) {
	if len(r.Hostname) < minHostnameLength {
		return false, "Hostname must be at least 3 characters"
	}
	if len(r.Hostname) > maxHostnameLength {
		return false, "Hostname must be less than 64 characters"
	}
	if !regexp.MustCompile(`^[A-Z0-9\-]+$`).MatchString(r.Hostname) {
		return false, "Hostname must be alphanumeric or -"
	}
	return true, ""
}

type TunnelWithPass struct {
	ID        uint      `json:"id"`
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Password  string    `json:"password"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}
