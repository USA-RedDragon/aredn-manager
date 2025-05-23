package v1

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/USA-RedDragon/aredn-manager/internal/services/olsr"
	"github.com/gin-gonic/gin"
)

func GETOLSRHosts(c *gin.Context) {
	olsrdParser, ok := c.MustGet("OLSRDHostParser").(*olsr.HostsParser)
	if !ok {
		slog.Error("GETOLSRHosts: OLSRDHostParser not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	pageStr, exists := c.GetQuery("page")
	if !exists {
		pageStr = "1"
	}
	pageInt, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		slog.Error("error parsing page", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page"})
		return
	}
	page := int(pageInt)

	limitStr, exists := c.GetQuery("limit")
	if !exists {
		limitStr = "50"
	}
	limitInt, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		slog.Error("Error parsing limit:", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	limit := int(limitInt)

	filter, exists := c.GetQuery("filter")
	if !exists {
		filter = ""
	}

	total := olsrdParser.GetHostsCount()

	nodes := olsrdParser.GetHostsPaginated(page, limit, filter)
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "total": total})
}

func GETOLSRHostsCount(c *gin.Context) {
	olsrdParser, ok := c.MustGet("OLSRDHostParser").(*olsr.HostsParser)
	if !ok {
		slog.Error("GETOLSRHostsCount: OLSRDHostParser not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	servicesParser, ok := c.MustGet("OLSRDServiceParser").(*olsr.ServicesParser)
	if !ok {
		slog.Error("GETOLSRHostsCount: OLSRDServiceParser not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": olsrdParser.GetAREDNHostsCount(), "total": olsrdParser.GetTotalHostsCount(), "services": servicesParser.GetServicesCount()})
}

func GETOLSRRunning(c *gin.Context) {
	registry, ok := c.MustGet("registry").(*services.Registry)
	if !ok {
		slog.Error("Error getting registry")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	olsrService, ok := registry.Get(services.OLSRServiceName)
	if !ok {
		slog.Error("Error getting OLSR service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"running": olsrService.IsRunning()})
}
