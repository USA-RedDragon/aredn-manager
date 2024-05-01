package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
)

func GETOLSRHosts(c *gin.Context) {
	olsrdParser, ok := c.MustGet("OLSRDHostParser").(*olsrd.HostsParser)
	if !ok {
		fmt.Println("POSTLogin: OLSRDHostParser not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	pageStr, exists := c.GetQuery("page")
	if !exists {
		pageStr = "1"
	}
	pageInt, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		fmt.Println("Error parsing page:", err)
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
		fmt.Println("Error parsing limit:", err)
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
	olsrdParser, ok := c.MustGet("OLSRDHostParser").(*olsrd.HostsParser)
	if !ok {
		fmt.Println("POSTLogin: OLSRDHostParser not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": olsrdParser.GetAREDNHostsCount(), "total": olsrdParser.GetTotalHostsCount()})
}

func GETOLSRRunning(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": olsrd.IsRunning()})
}
