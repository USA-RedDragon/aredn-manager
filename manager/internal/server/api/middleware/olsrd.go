package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
)

func OLSRDProvider(parsers *olsrd.Parsers) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("OLSRDParsers", parsers)
		c.Next()
	}
}
