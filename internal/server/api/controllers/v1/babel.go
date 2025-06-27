package v1

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/USA-RedDragon/aredn-manager/internal/server/api/middleware"
	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/gin-gonic/gin"
)

func GETBabelHosts(c *gin.Context) {
	di, ok := c.MustGet(middleware.DepInjectionKey).(*middleware.DepInjection)
	if !ok {
		slog.Error("Unable to get dependencies from context")
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

	total := di.AREDNLinkParser.GetHostsCount()

	nodes := di.AREDNLinkParser.GetHostsPaginated(page, limit, filter)
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "total": total})
}

func GETBabelHostsCount(c *gin.Context) {
	di, ok := c.MustGet(middleware.DepInjectionKey).(*middleware.DepInjection)
	if !ok {
		slog.Error("Unable to get dependencies from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes":    di.AREDNLinkParser.GetAREDNHostsCount(),
		"total":    di.AREDNLinkParser.GetTotalHostsCount(),
		"services": di.AREDNLinkParser.GetServiceCount(),
	})
}

func GETBabelRunning(c *gin.Context) {
	di, ok := c.MustGet(middleware.DepInjectionKey).(*middleware.DepInjection)
	if !ok {
		slog.Error("Unable to get dependencies from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	if !di.Config.Babel.Enabled {
		slog.Info("Babel service is not enabled in the configuration")
		c.JSON(http.StatusOK, gin.H{"running": false})
		return
	}

	babelService, ok := di.ServiceRegistry.Get(services.BabelServiceName)
	if !ok {
		slog.Error("Error getting Babel service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"running": babelService.IsRunning()})
}
