package v1

import (
	"log/slog"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/server/api/middleware"
	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/gin-gonic/gin"
)

func GETDNSRunning(c *gin.Context) {
	di, ok := c.MustGet(middleware.DepInjectionKey).(*middleware.DepInjection)
	if !ok {
		slog.Error("Unable to get dependencies from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	dnsmasqService, ok := di.ServiceRegistry.Get(services.DNSMasqServiceName)
	if !ok {
		slog.Error("Error getting DNSMasq service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"running": dnsmasqService.IsRunning()})
}
