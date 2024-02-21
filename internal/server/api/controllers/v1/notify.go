package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/dnsmasq"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error regenerating DNS"})
		fmt.Println("Error reloading DNS config:", err)
		return
	}

	go func() {
		olsrdParser, ok := c.MustGet("OLSRDHostParser").(*olsrd.HostsParser)
		if !ok {
			fmt.Println("POSTLogin: OLSRDHostParser not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
			return
		}
		err := olsrdParser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing hosts"})
			fmt.Println("Error parsing hosts:", err)
			return
		}

		olsrdServicesParser, ok := c.MustGet("OLSRDServicesParser").(*olsrd.ServicesParser)
		if !ok {
			fmt.Println("POSTLogin: OLSRDServicesParser not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
			return
		}

		err = olsrdServicesParser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing services"})
			fmt.Println("Error parsing services:", err)
			return
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
