package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/wireguard"
	"github.com/gin-gonic/gin"
)

func WireguardManagerProvider(manager *wireguard.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("WireguardManager", manager)
		c.Next()
	}
}
