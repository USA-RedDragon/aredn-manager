package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
)

func OLSRDProvider(parser *olsrd.HostsParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("OLSRDHostParser", parser)
		c.Next()
	}
}
