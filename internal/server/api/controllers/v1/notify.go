package v1

import (
	"log/slog"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/USA-RedDragon/aredn-manager/internal/services/arednlink"
	"github.com/USA-RedDragon/aredn-manager/internal/services/olsr"
	"github.com/gin-gonic/gin"
)

func POSTNotify(c *gin.Context) {
	if (c.RemoteIP() != "127.0.0.1" && c.RemoteIP() != "::1") || c.GetHeader("X-Forwarded-For") != "" {
		slog.Warn("POSTNotify: Forbidden notify", "ip", c.RemoteIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	registry, ok := c.MustGet("registry").(*services.Registry)
	if !ok {
		slog.Error("POSTNotify: Error getting registry")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	dnsmasqService, ok := registry.Get(services.DNSMasqServiceName)
	if !ok {
		slog.Error("POSTNotify: Error getting DNSMasq service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	err := dnsmasqService.Reload()
	if err != nil {
		slog.Error("POSTNotify: Error reloading DNSMasq service", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error regenerating DNS"})
		return
	}

	go func() {
		olsrdParser, ok := c.MustGet("OLSRDHostParser").(*olsr.HostsParser)
		if !ok {
			slog.Error("POSTNotify: OLSRDHostParser not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
			return
		}
		err := olsrdParser.Parse()
		if err != nil {
			slog.Error("POSTNotify: Error parsing hosts", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing hosts"})
			return
		}

		olsrdServicesParser, ok := c.MustGet("OLSRDServicesParser").(*olsr.ServicesParser)
		if !ok {
			slog.Error("POSTNotify: OLSRDServicesParser not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
			return
		}

		err = olsrdServicesParser.Parse()
		if err != nil {
			slog.Error("POSTNotify: Error parsing services", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing services"})
			return
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func POSTNotifyBabel(c *gin.Context) {
	if (c.RemoteIP() != "127.0.0.1" && c.RemoteIP() != "::1") || c.GetHeader("X-Forwarded-For") != "" {
		slog.Warn("POSTNotify: Forbidden notify", "ip", c.RemoteIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	go func() {
		arednlinkParser, ok := c.MustGet("AREDNLinkParser").(*arednlink.Parser)
		if !ok {
			slog.Error("POSTNotifyBabel: AREDNLinkParser not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
			return
		}
		err := arednlinkParser.Parse()
		if err != nil {
			slog.Error("POSTNotifyBabel: Error parsing", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing"})
			return
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
