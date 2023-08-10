package apimodels

import "regexp"

type CreateMesh struct {
	Name string   `json:"name" binding:"required"`
	IPs  []string `json:"ips" binding:"required"`
}

func (r *CreateMesh) IsValidHostname() (bool, string) {
	if len(r.Name) < minHostnameLength {
		return false, "Name must be at least 3 characters"
	}
	if len(r.Name) > maxHostnameLength {
		return false, "Name must be less than 64 characters"
	}
	if !regexp.MustCompile(`^[A-Z0-9\-]+$`).MatchString(r.Name) {
		return false, "Name must be alphanumeric or -"
	}
	return true, ""
}
