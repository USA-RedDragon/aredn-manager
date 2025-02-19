package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/services/olsr"
	"github.com/gin-gonic/gin"
)

func OLSRDServicesProvider(parser *olsr.ServicesParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("OLSRDServicesParser", parser)
		c.Next()
	}
}
