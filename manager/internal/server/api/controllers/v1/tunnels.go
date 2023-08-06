package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/USA-RedDragon/aredn-manager/internal/vtun"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GETTunnels(c *gin.Context) {
	db, ok := c.MustGet("PaginatedDB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	cDb, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	tunnels, err := models.ListTunnels(db)
	if err != nil {
		fmt.Printf("GETTunnels: Error getting tunnels: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnels"})
		return
	}

	total, err := models.CountTunnels(cDb)
	if err != nil {
		fmt.Printf("GETTunnels: Error getting tunnel count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
		return
	}

	adminStr, exists := c.GetQuery("admin")
	if !exists {
		adminStr = "false"
	}
	admin, err := strconv.ParseBool(adminStr)
	if err != nil {
		fmt.Printf("GETTunnels: Error parsing admin query: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing admin query"})
		return
	}

	if admin {
		// Check for an active session
		session := sessions.Default(c)
		user := session.Get("user_id")
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Change the json response to include the password
		var tunnelsWithPass []apimodels.TunnelWithPass

		for _, tunnel := range tunnels {
			tunnelsWithPass = append(tunnelsWithPass, apimodels.TunnelWithPass{
				ID:        tunnel.ID,
				Hostname:  tunnel.Hostname,
				IP:        tunnel.IP,
				Password:  tunnel.Password,
				Active:    tunnel.Active,
				CreatedAt: tunnel.CreatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"total": total, "tunnels": tunnelsWithPass})
	} else {
		c.JSON(http.StatusOK, gin.H{"total": total, "tunnels": tunnels})
	}
}

func POSTTunnel(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("POSTUser: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.CreateTunnel
	err := c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("POSTTunnel: JSON data is invalid: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		isValid, errString := json.IsValidHostname()
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": errString})
			return
		}

		if json.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password cannot be empty"})
			return
		}

		// Check if the hostname is already taken
		var tunnel models.Tunnel
		err := db.Find(&tunnel, "hostname = ?", json.Hostname).Error
		if err != nil {
			fmt.Printf("POSTTunnel: Error getting tunnel: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel"})
			return
		} else if tunnel.ID != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hostname is already taken"})
			return
		}

		tunnel = models.Tunnel{
			Hostname: json.Hostname,
			Password: json.Password,
		}
		tunnel.IP, err = models.GetNextIP(db)
		if err != nil {
			fmt.Printf("POSTTunnel: Error getting next IP: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting next IP"})
			return
		}

		err = db.Create(&tunnel).Error
		if err != nil {
			fmt.Printf("POSTTunnel: Error creating tunnel: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating tunnel"})
			return
		}

		vtun.GenerateAndSave(config, db)
		if err != nil {
			fmt.Printf("POSTTunnel: Error generating vtun config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun config"})
			return
		}

		olsrd.GenerateAndSave(config, db)
		if err != nil {
			fmt.Printf("POSTTunnel: Error generating olsrd config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating olsrd config"})
			return
		}

		err = vtun.Reload()
		if err != nil {
			fmt.Printf("POSTTunnel: Error reloading vtun: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun"})
			return
		}

		err = olsrd.Reload()
		if err != nil {
			fmt.Printf("POSTTunnel: Error reloading olsrd: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading olsrd"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tunnel created"})
	}
}

func DELETETunnel(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("POSTUser: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	idUint64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tunnel ID"})
		return
	}

	exists, err := models.TunnelIDExists(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error checking if tunnel exists: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking if tunnel exists"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tunnel does not exist"})
		return
	}

	err = models.DeleteTunnel(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error deleting tunnel: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting tunnel"})
		return
	}

	vtun.GenerateAndSave(config, db)
	if err != nil {
		fmt.Printf("Error generating vtun config: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun config"})
		return
	}

	olsrd.GenerateAndSave(config, db)
	if err != nil {
		fmt.Printf("Error generating olsrd config: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating olsrd config"})
		return
	}

	err = vtun.Reload()
	if err != nil {
		fmt.Printf("Error reloading vtun: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun"})
		return
	}

	err = olsrd.Reload()
	if err != nil {
		fmt.Printf("Error reloading olsrd: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading olsrd"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tunnel deleted"})
}
