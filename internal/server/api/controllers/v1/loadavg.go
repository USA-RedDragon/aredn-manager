package v1

import (
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"github.com/gin-gonic/gin"
)

func GETLoadAvg(c *gin.Context) {
	loadavg, err := utils.GetLoadAvg()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get load average"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"loadavg": loadavg})
}
