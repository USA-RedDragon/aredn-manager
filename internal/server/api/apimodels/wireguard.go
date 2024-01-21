package apimodels

type WireguardPubkeyRequest struct {
	Privkey string `json:"key" binding:"required"`
}
