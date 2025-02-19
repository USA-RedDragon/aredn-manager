package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/gin-gonic/gin"
)

func ServiceRegistryProvider(registry *services.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("registry", registry)
		c.Next()
	}
}
