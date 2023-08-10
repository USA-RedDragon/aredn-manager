package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/bind"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
)

func POSTNotify(c *gin.Context) {
	if (c.RemoteIP() != "127.0.0.1" && c.RemoteIP() != "::1") || c.GetHeader("X-Forwarded-For") != "" {
		fmt.Println("Forbidden notify from", c.RemoteIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	olsrdParsers, ok := c.MustGet("OLSRDParsers").(*olsrd.Parsers)
	if !ok {
		fmt.Println("POSTLogin: OLSRDParsers not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	err := bind.Reload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error notifying bind"})
		fmt.Println("Error notifying bind:", err)
		return
	}
	err = olsrdParsers.HostsParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing hosts"})
		fmt.Println("Error parsing hosts:", err)
		return
	}
	err = olsrdParsers.ServicesParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing services"})
		fmt.Println("Error parsing services:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
