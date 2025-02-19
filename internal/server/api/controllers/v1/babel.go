package v1

import (
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/babel"
	"github.com/gin-gonic/gin"
)

func GETBabelRunning(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": babel.IsRunning()})
}
