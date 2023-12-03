package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
)

func OLSRDServicesProvider(parser *olsrd.ServicesParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("OLSRDServicesParser", parser)
		c.Next()
	}
}
