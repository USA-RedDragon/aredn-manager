package v1

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/dnsmasq"
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
				ID:             tunnel.ID,
				Hostname:       tunnel.Hostname,
				IP:             tunnel.IP,
				Password:       tunnel.Password,
				Client:         tunnel.Client,
				Active:         tunnel.Active,
				ConnectionTime: tunnel.ConnectionTime,
				CreatedAt:      tunnel.CreatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"total": total, "tunnels": tunnelsWithPass})
	} else {
		c.JSON(http.StatusOK, gin.H{"total": total, "tunnels": tunnels})
	}
}

func GETTunnelsCount(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	count, err := models.CountTunnels(db)
	if err != nil {
		fmt.Printf("GETTunnelsCount: Error getting tunnel count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func GETTunnelsCountConnected(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	count, err := models.CountActiveTunnels(db)
	if err != nil {
		fmt.Printf("GETTunnelsCountConnected: Error getting tunnel count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
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
		fmt.Println("POSTTunnel: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	vtunClientWatcher, ok := c.MustGet("VTunClientWatcher").(*vtun.VTunClientWatcher)
	if !ok {
		fmt.Println("DELETETunnel: Unable to get VTunClientWatcher from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.CreateTunnel
	err := c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("POSTTunnel: JSON data is invalid: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		if json.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password cannot be empty"})
			return
		}

		if !json.Client {
			json.Hostname = strings.ToUpper(json.Hostname)
			isValid, errString := json.IsValidHostname()
			if !isValid {
				c.JSON(http.StatusBadRequest, gin.H{"error": errString})
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
				Client:   json.Client,
			}
			tunnel.IP, err = models.GetNextIP(db, config)
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

			err = vtun.GenerateAndSave(config, db)
			if err != nil {
				fmt.Printf("POSTTunnel: Error generating vtun config: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun config"})
				return
			}

			err = vtun.Reload()
			if err != nil {
				fmt.Printf("POSTTunnel: Error reloading vtun: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun"})
				return
			}

		} else {
			if json.IP == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP cannot be empty"})
				return
			}

			// Check to ensure the IP is valid
			if net.ParseIP(json.IP) == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP is not a valid IP address"})
				return
			}

			// Check to ensure the IP is in the correct range: 172.16.0.0/12
			ip := net.ParseIP(json.IP)
			_, cidr, err := net.ParseCIDR("172.16.0.0/12")
			if err != nil {
				fmt.Printf("POSTTunnel: Error parsing CIDR: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing CIDR"})
				return
			}
			if !cidr.Contains(ip) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP is not in the correct range"})
				return
			}

			// Check to ensure the hostname is either a valid IP or a valid address (without protocol) with an optional port

			// split the hostname by :
			// if len(split) == 1, then it's just an IP address
			// if len(split) == 2, then it's an address with an optional port
			// if len(split) > 2, then it's invalid

			split := strings.Split(json.Hostname, ":")
			if len(split) > 2 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Server address is invalid"})
				return
			}

			// Check if the hostname is an IP address
			if net.ParseIP(split[0]) == nil {
				// Check if the hostname is a valid address
				_, err := url.ParseRequestURI("http://" + split[0])
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Server address is invalid"})
					return
				}

				// Check that the hostname is resolvable
				_, err = net.LookupIP(split[0])
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Server address is not resolvable"})
					return
				}
			}

			// Check if the port is valid
			if len(split) == 2 {
				port, err := strconv.ParseUint(split[1], 10, 16)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Server port is invalid"})
					return
				}
				if port < 1 || port > 65535 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Server port is invalid"})
					return
				}
			}

			// Check if the IP is already taken
			var tunnel models.Tunnel
			err = db.Find(&tunnel, "ip = ?", json.IP).Error
			if err != nil {
				fmt.Printf("POSTTunnel: Error getting tunnel: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel"})
				return
			} else if tunnel.ID != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is already taken"})
				return
			}

			tunnel = models.Tunnel{
				Hostname: json.Hostname,
				Password: json.Password,
				IP:       json.IP,
				Client:   json.Client,
			}

			err = db.Create(&tunnel).Error
			if err != nil {
				fmt.Printf("POSTTunnel: Error creating tunnel: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating tunnel"})
				return
			}

			err = vtun.GenerateAndSaveClient(config, db)
			if err != nil {
				fmt.Printf("POSTTunnel: Error generating vtun client config: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun client config"})
				return
			}

			err = vtun.ReloadAllClients(db, vtunClientWatcher)
			if err != nil {
				fmt.Printf("POSTTunnel: Error reloading vtun client: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun client"})
				return
			}
		}

		err = olsrd.GenerateAndSave(config, db)
		if err != nil {
			fmt.Printf("POSTTunnel: Error generating olsrd config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating olsrd config"})
			return
		}

		err = olsrd.Reload()
		if err != nil {
			fmt.Printf("POSTTunnel: Error reloading olsrd: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading olsrd"})
			return
		}

		err = dnsmasq.Reload()
		if err != nil {
			fmt.Printf("Error reloading DNS: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading DNS"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tunnel created"})
	}
}

func PATCHTunnel(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("PATCHTunnel: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	vtunClientWatcher, ok := c.MustGet("VTunClientWatcher").(*vtun.VTunClientWatcher)
	if !ok {
		fmt.Println("PATCHTunnel: Unable to get VTunClientWatcher from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.EditTunnel
	err := c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("PATCHTunnel: JSON data is invalid: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		exists, err := models.TunnelIDExists(db, json.ID)
		if err != nil {
			fmt.Printf("Error checking if tunnel exists: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking if tunnel exists"})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tunnel does not exist"})
			return
		}
		if json.IP == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IP cannot be empty"})
			return
		}

		// Check to ensure the IP is valid
		if net.ParseIP(json.IP) == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IP is not a valid IP address"})
			return
		}

		// Check to ensure the IP is in the correct range: 172.16.0.0/12
		ip := net.ParseIP(json.IP)
		_, cidr, err := net.ParseCIDR("172.16.0.0/12")
		if err != nil {
			fmt.Printf("PATCHTunnel: Error parsing CIDR: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing CIDR"})
			return
		}
		if !cidr.Contains(ip) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IP is not in the correct range"})
			return
		}

		// Check to ensure the hostname is either a valid IP or a valid address (without protocol) with an optional port

		// split the hostname by :
		// if len(split) == 1, then it's just an IP address
		// if len(split) == 2, then it's an address with an optional port
		// if len(split) > 2, then it's invalid

		json.Hostname = strings.ToUpper(json.Hostname)

		split := strings.Split(json.Hostname, ":")
		if len(split) > 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server address is invalid"})
			return
		}

		// Check if the hostname is an IP address
		if net.ParseIP(split[0]) == nil {
			// Check if the hostname is a valid address
			_, err := url.ParseRequestURI("http://" + split[0])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Server address is invalid"})
				return
			}

			// Check that the hostname is resolvable
			_, err = net.LookupIP(split[0])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Server address is not resolvable"})
				return
			}
		}

		// Check if the port is valid
		if len(split) == 2 {
			port, err := strconv.ParseUint(split[1], 10, 16)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Server port is invalid"})
				return
			}
			if port < 1 || port > 65535 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Server port is invalid"})
				return
			}
		}

		// Check if the IP is already taken
		var tunnel models.Tunnel
		err = db.Find(&tunnel, "ip = ?", json.IP).Error
		if err != nil {
			fmt.Printf("PATCHTunnel: Error getting tunnel: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel"})
			return
		} else if tunnel.ID != 0 && tunnel.ID != json.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is already taken"})
			return
		}

		tunnel.Hostname = json.Hostname
		tunnel.Password = json.Password
		tunnel.IP = json.IP

		err = db.Save(&tunnel).Error
		if err != nil {
			fmt.Printf("PATCHTunnel: Error saving tunnel: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving tunnel"})
			return
		}

		err = vtun.GenerateAndSaveClient(config, db)
		if err != nil {
			fmt.Printf("PATCHTunnel: Error generating vtun client config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun client config"})
			return
		}

		err = vtun.ReloadAllClients(db, vtunClientWatcher)
		if err != nil {
			fmt.Printf("PATCHTunnel: Error reloading vtun client: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun client"})
			return
		}

		err = olsrd.GenerateAndSave(config, db)
		if err != nil {
			fmt.Printf("PATCHTunnel: Error generating olsrd config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating olsrd config"})
			return
		}

		err = olsrd.Reload()
		if err != nil {
			fmt.Printf("PATCHTunnel: Error reloading olsrd: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading olsrd"})
			return
		}

		err = dnsmasq.Reload()
		if err != nil {
			fmt.Printf("Error reloading DNS: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading DNS"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tunnel updated"})
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
		fmt.Println("DELETETunnel: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	vtunClientWatcher, ok := c.MustGet("VTunClientWatcher").(*vtun.VTunClientWatcher)
	if !ok {
		fmt.Println("DELETETunnel: Unable to get VTunClientWatcher from context")
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

	err = vtun.GenerateAndSave(config, db)
	if err != nil {
		fmt.Printf("Error generating vtun config: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun config"})
		return
	}

	err = vtun.GenerateAndSaveClient(config, db)
	if err != nil {
		fmt.Printf("DELETETunnel: Error generating vtun client config: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun client config"})
		return
	}

	err = olsrd.GenerateAndSave(config, db)
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

	err = vtun.ReloadAllClients(db, vtunClientWatcher)
	if err != nil {
		fmt.Printf("DELETETunnel: Error reloading vtun client: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun client"})
		return
	}

	err = olsrd.Reload()
	if err != nil {
		fmt.Printf("Error reloading olsrd: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading olsrd"})
		return
	}

	err = dnsmasq.Reload()
	if err != nil {
		fmt.Printf("Error reloading DNS: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading DNS"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tunnel deleted"})
}
