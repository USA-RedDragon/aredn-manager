package middleware

import (
	"github.com/USA-RedDragon/mesh-manager/internal/bandwidth"
	"github.com/USA-RedDragon/mesh-manager/internal/config"
	"github.com/USA-RedDragon/mesh-manager/internal/services"
	"github.com/USA-RedDragon/mesh-manager/internal/services/arednlink"
	"github.com/USA-RedDragon/mesh-manager/internal/services/olsr"
	"github.com/USA-RedDragon/mesh-manager/internal/wireguard"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DepInjection struct {
	AREDNLinkParser    *arednlink.Parser
	Config             *config.Config
	DB                 *gorm.DB
	PaginatedDB        *gorm.DB
	NetworkStats       *bandwidth.StatCounterManager
	OLSRHostsParser    *olsr.HostsParser
	OLSRServicesParser *olsr.ServicesParser
	ServiceRegistry    *services.Registry
	Version            string
	WireguardManager   *wireguard.Manager
}

const DepInjectionKey = "DepInjection"

func Inject(inj *DepInjection) gin.HandlerFunc {
	return func(c *gin.Context) {
		inj.DB = inj.DB.WithContext(c.Request.Context())
		inj.PaginatedDB = paginateDB(inj.DB, c)

		c.Set(DepInjectionKey, inj)
		c.Next()
	}
}
