package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/gin-gonic/gin"
)

func GETBabelRunning(c *gin.Context) {
	registry, ok := c.MustGet("registry").(*services.Registry)
	if !ok {
		fmt.Println("Error getting registry")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	babelService, ok := registry.Get(services.BabelServiceName)
	if !ok {
		fmt.Println("Error getting Babel service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"running": babelService.IsRunning()})
}
