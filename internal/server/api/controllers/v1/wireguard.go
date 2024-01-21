package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/gin-gonic/gin"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func GETWireguardGenkey(c *gin.Context) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating key"})
		fmt.Println("Error generating key:", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key.String()})
}

func POSTWireguardPubkey(c *gin.Context) {
	var req apimodels.WireguardPubkeyRequest
	err := c.BindJSON(&req)
	if err != nil {
		fmt.Println("Error binding json:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	private, err := wgtypes.ParseKey(req.Privkey)
	if err != nil {
		fmt.Println("Error parsing key:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": private.PublicKey().String()})
}
