package v1

import (
	"log/slog"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/bandwidth"
	"github.com/gin-gonic/gin"
)

func GETStats(c *gin.Context) {
	stats, ok := c.MustGet("NetworkStats").(*bandwidth.StatCounterManager)
	if !ok {
		slog.Error("GETStats: Unable to get stats manager from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	statsJSON := gin.H{
		"total_rx_mb":            stats.TotalRXMB,
		"total_tx_mb":            stats.TotalTXMB,
		"total_rx_bytes_per_sec": stats.TotalRXBandwidth,
		"total_tx_bytes_per_sec": stats.TotalTXBandwidth,
	}
	c.JSON(http.StatusOK, gin.H{"stats": statsJSON})
}
