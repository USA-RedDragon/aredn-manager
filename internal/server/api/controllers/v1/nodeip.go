package v1

import (
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/gin-gonic/gin"
)

func GETNodeIP(c *gin.Context) {
	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodeIP": config.NodeIP})
}
