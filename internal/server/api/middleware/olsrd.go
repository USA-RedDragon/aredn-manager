package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/services/olsr"
	"github.com/gin-gonic/gin"
)

func OLSRDProvider(parser *olsr.HostsParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("OLSRDHostParser", parser)
		c.Next()
	}
}
