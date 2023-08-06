package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/gin-gonic/gin"
)

func ConfigProvider(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Config", config)
		c.Next()
	}
}
