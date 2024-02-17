package middleware

import (
	"github.com/gin-gonic/gin"
)

func VersionProvider(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Version", version)
		c.Next()
	}
}
