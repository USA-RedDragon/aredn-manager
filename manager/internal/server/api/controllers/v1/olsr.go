package v1

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/gin-gonic/gin"
)

func GETOLSRHosts(c *gin.Context) {
	olsrdParsers, ok := c.MustGet("OLSRDParsers").(*olsrd.Parsers)
	if !ok {
		fmt.Println("POSTLogin: OLSRDParsers not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	nodes := olsrdParsers.HostsParser.GetHosts()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

func GETOLSRServices(c *gin.Context) {
	olsrdParsers, ok := c.MustGet("OLSRDParsers").(*olsrd.Parsers)
	if !ok {
		fmt.Println("POSTLogin: OLSRDParsers not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	services := olsrdParsers.ServicesParser.GetServices()
	c.JSON(http.StatusOK, gin.H{"services": services})
}

func GETOLSRRunning(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": olsrd.IsRunning()})
}
