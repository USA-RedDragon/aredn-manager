package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/bandwidth"
	"github.com/gin-gonic/gin"
)

func NetworkStats(stats *bandwidth.StatCounterManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("NetworkStats", stats)
		c.Next()
	}
}
