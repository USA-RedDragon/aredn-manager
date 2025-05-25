package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/services/arednlink"
	"github.com/gin-gonic/gin"
)

func AREDNLinkParserProvider(parser *arednlink.Parser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("AREDNLinkParser", parser)
		c.Next()
	}
}
