package v1

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GETVersion(c *gin.Context) {
	version, ok := c.MustGet("Version").(string)
	if !ok {
		slog.Error("GETVersion: Unable to get version from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	_, err := io.WriteString(c.Writer, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting version"})
	}
}

func GETPing(c *gin.Context) {
	_, err := io.WriteString(c.Writer, fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting ping"})
	}
}
