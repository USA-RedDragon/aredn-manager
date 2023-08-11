package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
)

func GETOLSRHosts(c *gin.Context) {
	olsrdParser, ok := c.MustGet("OLSRDHostParser").(*olsrd.HostsParser)
	if !ok {
		fmt.Println("POSTLogin: OLSRDHostParser not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	nodes := olsrdParser.GetHosts()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

func GETOLSRRunning(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": olsrd.IsRunning()})
}
