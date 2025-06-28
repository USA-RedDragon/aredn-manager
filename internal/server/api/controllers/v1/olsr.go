package v1

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/USA-RedDragon/mesh-manager/internal/server/api/middleware"
	"github.com/USA-RedDragon/mesh-manager/internal/services"
	"github.com/gin-gonic/gin"
)

func GETOLSRHosts(c *gin.Context) {
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

	total := di.OLSRHostsParser.GetHostsCount()

	nodes := di.OLSRHostsParser.GetHostsPaginated(page, limit, filter)
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "total": total})
}

func GETOLSRHostsCount(c *gin.Context) {
	di, ok := c.MustGet(middleware.DepInjectionKey).(*middleware.DepInjection)
	if !ok {
		slog.Error("Unable to get dependencies from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes":    di.OLSRHostsParser.GetMeshHostsCount(),
		"total":    di.OLSRHostsParser.GetTotalHostsCount(),
		"services": di.OLSRServicesParser.GetServicesCount(),
	})
}

func GETOLSRRunning(c *gin.Context) {
	di, ok := c.MustGet(middleware.DepInjectionKey).(*middleware.DepInjection)
	if !ok {
		slog.Error("Unable to get dependencies from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	if !di.Config.OLSR {
		c.JSON(http.StatusOK, gin.H{"running": false})
		return
	}

	olsrService, ok := di.ServiceRegistry.Get(services.OLSRServiceName)
	if !ok {
		slog.Error("Error getting OLSR service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"running": olsrService.IsRunning()})
}
