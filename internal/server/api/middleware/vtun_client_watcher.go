package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/vtun"
	"github.com/gin-gonic/gin"
)

func VTunClientWatcherProvider(clientWatcher *vtun.VTunClientWatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("VTunClientWatcher", clientWatcher)
		c.Next()
	}
}
