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
	"github.com/USA-RedDragon/aredn-manager/internal/wireguard"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

	typeStr, exists := c.GetQuery("type")
	if !exists {
		typeStr = "vtun"
	}

	var tunnels []models.Tunnel
	var total int
	switch typeStr {
	case "vtun":
		var err error
		tunnels, err = models.ListVtunTunnels(db)
		if err != nil {
			fmt.Printf("GETTunnels: Error getting tunnels: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnels"})
			return
		}
		total, err = models.CountVtunTunnels(cDb)
		if err != nil {
			fmt.Printf("GETTunnels: Error getting tunnel count: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
			return
		}
	case "wireguard":
		var err error
		tunnels, err = models.ListWireguardTunnels(db)
		if err != nil {
			fmt.Printf("GETTunnels: Error getting tunnels: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnels"})
			return
		}
		total, err = models.CountWireguardTunnels(cDb)
		if err != nil {
			fmt.Printf("GETTunnels: Error getting tunnel count: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
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
				Wireguard:      tunnel.Wireguard,
				WireguardPort:  tunnel.WireguardPort,
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

func GETVTunTunnelsCount(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	count, err := models.CountVtunTunnels(db)
	if err != nil {
		fmt.Printf("GETVTunTunnelsCount: Error getting tunnel count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func GETVTunTunnelsCountConnected(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	count, err := models.CountVTunActiveTunnels(db)
	if err != nil {
		fmt.Printf("GETTunnelsCountConnected: Error getting tunnel count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func GETWireguardTunnelsCount(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	count, err := models.CountWireguardTunnels(db)
	if err != nil {
		fmt.Printf("GETWireguardTunnelsCount: Error getting tunnel count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func GETWireguardTunnelsCountConnected(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	count, err := models.CountWireguardActiveTunnels(db)
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

	wireguardManager, ok := c.MustGet("WireguardManager").(*wireguard.Manager)
	if !ok {
		fmt.Println("POSTTunnel: Unable to get WireguardManager from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.CreateTunnel
	err := c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("POSTTunnel: JSON data is invalid: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		if (!json.Wireguard || json.Client) && json.Password == "" {
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
			err := db.Find(&tunnel, "hostname = ? AND wireguard = ?", json.Hostname, json.Wireguard).Error
			if err != nil {
				fmt.Printf("POSTTunnel: Error getting tunnel: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel"})
				return
			} else if tunnel.ID != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Hostname is already taken"})
				return
			}

			tunnel = models.Tunnel{
				Hostname:  json.Hostname,
				Password:  json.Password,
				Client:    json.Client,
				Wireguard: json.Wireguard,
			}
			if !tunnel.Wireguard {
				tunnel.IP, err = models.GetNextVTunIP(db, config)
				if err != nil {
					fmt.Printf("POSTTunnel: Error getting next IP: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting next IP"})
					return
				}
			} else {
				tunnel.IP, err = models.GetNextWireguardIP(db, config)
				if err != nil {
					fmt.Printf("POSTTunnel: Error getting next IP: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting next IP"})
					return
				}

				tunnel.WireguardPort, err = models.GetNextWireguardPort(db, config)
				if err != nil {
					fmt.Printf("POSTTunnel: Error getting next port: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting next port"})
					return
				}

				// Generate a server and client key pair
				serverKey, err := wgtypes.GeneratePrivateKey()
				if err != nil {
					fmt.Printf("POSTTunnel: Error generating server key: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating server key"})
					return
				}
				clientKey, err := wgtypes.GeneratePrivateKey()
				if err != nil {
					fmt.Printf("POSTTunnel: Error generating client key: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating client key"})
					return
				}

				tunnel.Password = serverKey.PublicKey().String() + clientKey.String() + clientKey.PublicKey().String()
				tunnel.WireguardServerKey = serverKey.String()
				tunnel.TunnelInterface = wireguard.GenerateWireguardInterfaceName(tunnel)
			}

			err = db.Create(&tunnel).Error
			if err != nil {
				fmt.Printf("POSTTunnel: Error creating tunnel: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating tunnel"})
				return
			}

			if !tunnel.Wireguard {
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
				err = wireguardManager.AddPeer(tunnel)
				if err != nil {
					fmt.Printf("POSTTunnel: Error adding wireguard peer: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding wireguard peer"})
					return
				}
			}
		} else {
			if json.IP == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP cannot be empty"})
				return
			}

			if json.Wireguard {
				// json.Hostname must not contain a port
				if strings.Contains(json.Hostname, ":") {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Server address is invalid"})
					return
				}

				// json.IP must contain a port that needs to be appended to the hostname instead
				ipParts := strings.Split(json.IP, ":")
				if len(ipParts) != 2 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Net is invalid"})
					return
				}
				json.Hostname = json.Hostname + ":" + ipParts[1]
				json.IP = ipParts[0]
			}

			// Check to ensure the IP is in the correct range: 172.16.0.0/12
			ip := net.ParseIP(json.IP)
			if ip == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP is not a valid IP address"})
				return
			}
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
				Hostname:  json.Hostname,
				Password:  json.Password,
				IP:        json.IP,
				Client:    json.Client,
				Wireguard: json.Wireguard,
			}

			if tunnel.Wireguard {
				tunnel.TunnelInterface = wireguard.GenerateWireguardInterfaceName(tunnel)

				// The password will be 3 wireguard keys concatenated together
				// <server_pubkey><client_privkey><client_pubkey>
				if len(json.Password) != 132 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Key is invalid"})
					return
				}
				serverPubkey := json.Password[:44]
				_, err := wgtypes.ParseKey(serverPubkey)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Key is invalid"})
					return
				}
				clientPrivkey := json.Password[44:88]
				_, err = wgtypes.ParseKey(clientPrivkey)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Key is invalid"})
					return
				}
				clientPubkey := json.Password[88:]
				_, err = wgtypes.ParseKey(clientPubkey)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Key is invalid"})
					return
				}
			}

			err = db.Create(&tunnel).Error
			if err != nil {
				fmt.Printf("POSTTunnel: Error creating tunnel: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating tunnel"})
				return
			}

			if !tunnel.Wireguard {
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
			} else {
				err = wireguardManager.AddPeer(tunnel)
				if err != nil {
					fmt.Printf("POSTTunnel: Error adding wireguard peer: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding wireguard peer"})
					return
				}
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

	wireguardManager, ok := c.MustGet("WireguardManager").(*wireguard.Manager)
	if !ok {
		fmt.Println("POSTTunnel: Unable to get WireguardManager from context")
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

		origTunnel := tunnel

		tunnel.Hostname = json.Hostname
		tunnel.Password = json.Password
		tunnel.IP = json.IP

		if tunnel.Wireguard != json.Wireguard {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Changing tunnel type not allowed"})
			return
		}

		err = db.Save(&tunnel).Error
		if err != nil {
			fmt.Printf("PATCHTunnel: Error saving tunnel: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving tunnel"})
			return
		}

		if !tunnel.Wireguard {
			err = vtun.GenerateAndSave(config, db)
			if err != nil {
				fmt.Printf("PATCHTunnel: Error generating vtun config: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun config"})
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
		} else {
			err = wireguardManager.RemovePeer(origTunnel)
			if err != nil {
				fmt.Printf("PATCHTunnel: Error adding wireguard peer: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding wireguard peer"})
				return
			}

			err = wireguardManager.AddPeer(tunnel)
			if err != nil {
				fmt.Printf("PATCHTunnel: Error adding wireguard peer: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding wireguard peer"})
				return
			}
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

	wireguardManager, ok := c.MustGet("WireguardManager").(*wireguard.Manager)
	if !ok {
		fmt.Println("DELETETunnel: Unable to get WireguardManager from context")
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

	tunnel, err := models.FindTunnelByID(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error getting tunnel: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tunnel"})
		return
	}

	err = models.DeleteTunnel(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error deleting tunnel: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting tunnel"})
		return
	}

	if !tunnel.Wireguard {
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
	} else {
		err = wireguardManager.RemovePeer(tunnel)
		if err != nil {
			fmt.Printf("DELETETunnel: Error removing wireguard peer: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing wireguard peer"})
			return
		}
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
