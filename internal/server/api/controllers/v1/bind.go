package v1

import (
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/bind"
	"github.com/gin-gonic/gin"
)

func GETBindRunning(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": bind.IsRunning()})
}
