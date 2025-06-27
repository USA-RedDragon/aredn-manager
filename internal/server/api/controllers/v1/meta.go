package v1

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/server/api/middleware"
	"github.com/gin-gonic/gin"
)

func GETVersion(c *gin.Context) {
	di, ok := c.MustGet(middleware.DepInjectionKey).(*middleware.DepInjection)
	if !ok {
		slog.Error("Unable to get dependencies from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	_, err := io.WriteString(c.Writer, di.Version)
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
