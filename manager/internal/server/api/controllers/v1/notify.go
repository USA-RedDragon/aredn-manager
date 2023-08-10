package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/bind"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func POSTNotify(c *gin.Context) {
	if (c.RemoteIP() != "127.0.0.1" && c.RemoteIP() != "::1") || c.GetHeader("X-Forwarded-For") != "" {
		fmt.Println("Forbidden notify from", c.RemoteIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("POSTLogin: Unable to get DB from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("POSTLogin: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	olsrdParsers, ok := c.MustGet("OLSRDParsers").(*olsrd.Parsers)
	if !ok {
		fmt.Println("POSTLogin: OLSRDParsers not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	err := olsrdParsers.HostsParser.Parse()
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
	err = bind.GenerateAndSave(config, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating bind config"})
		fmt.Println("Error generating bind config:", err)
		return
	}

	err = bind.Reload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error notifying bind"})
		fmt.Println("Error notifying bind:", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
