package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/services/vtun"
	"github.com/gin-gonic/gin"
)

func VTunClientWatcherProvider(clientWatcher *vtun.ClientWatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("VTunClientWatcher", clientWatcher)
		c.Next()
	}
}
