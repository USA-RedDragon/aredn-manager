package v1

import (
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/dnsmasq"
	"github.com/gin-gonic/gin"
)

func GETDNSRunning(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": dnsmasq.IsRunning()})
}
