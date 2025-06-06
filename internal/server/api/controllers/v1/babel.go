package v1

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/USA-RedDragon/aredn-manager/internal/services/arednlink"
	"github.com/gin-gonic/gin"
)

func GETBabelHosts(c *gin.Context) {
	arednlinkParser, ok := c.MustGet("AREDNLinkParser").(*arednlink.Parser)
	if !ok {
		slog.Error("GETBabelHosts: AREDNLinkParser not found in context")
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

	total := arednlinkParser.GetHostsCount()

	nodes := arednlinkParser.GetHostsPaginated(page, limit, filter)
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "total": total})
}

func GETBabelHostsCount(c *gin.Context) {
	arednlinkParser, ok := c.MustGet("AREDNLinkParser").(*arednlink.Parser)
	if !ok {
		slog.Error("GETBabelHostsCount: AREDNLinkParser not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": arednlinkParser.GetAREDNHostsCount(), "total": arednlinkParser.GetTotalHostsCount(), "services": arednlinkParser.GetServiceCount()})
}

func GETBabelRunning(c *gin.Context) {
	registry, ok := c.MustGet("registry").(*services.Registry)
	if !ok {
		slog.Error("Error getting registry")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		slog.Error("Error getting config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	if !config.Babel.Enabled {
		slog.Info("Babel service is not enabled in the configuration")
		c.JSON(http.StatusOK, gin.H{"running": false})
		return
	}

	babelService, ok := registry.Get(services.BabelServiceName)
	if !ok {
		slog.Error("Error getting Babel service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"running": babelService.IsRunning()})
}
