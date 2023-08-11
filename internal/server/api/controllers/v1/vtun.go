package v1

import (
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/vtun"
	"github.com/gin-gonic/gin"
)

func GETVtunRunning(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": vtun.IsRunning()})
}
