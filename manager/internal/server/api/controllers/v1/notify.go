package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/dnsmasq"
	"github.com/gin-gonic/gin"
)

func POSTNotify(c *gin.Context) {
	if (c.RemoteIP() != "127.0.0.1" && c.RemoteIP() != "::1") || c.GetHeader("X-Forwarded-For") != "" {
		fmt.Println("Forbidden notify from", c.RemoteIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	err := dnsmasq.Reload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error notifying dnsmasq"})
		fmt.Println("Error notifying dnsmasq:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
