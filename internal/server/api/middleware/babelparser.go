package middleware

import (
	"github.com/USA-RedDragon/aredn-manager/internal/services/babel"
	"github.com/gin-gonic/gin"
)

func BabelParserProvider(parser *babel.Parser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("BabelParser", parser)
		c.Next()
	}
}
