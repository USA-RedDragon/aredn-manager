package v1

import (
	"fmt"
	"net/http"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"github.com/gin-gonic/gin"
)

func GETUptime(c *gin.Context) {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		fmt.Printf("GETUptime: Unable to get system info: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get system info"})
		return
	}
	uptime := utils.SecondsToClock(info.Uptime)
	if uptime == "" {
		fmt.Println("GETUptime: Unable to convert uptime to string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to convert uptime to string"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uptime": uptime})
}
