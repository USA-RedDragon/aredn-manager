package api

import (
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/events"
	v1Controllers "github.com/USA-RedDragon/aredn-manager/internal/server/api/controllers/v1"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/middleware"
	websocketControllers "github.com/USA-RedDragon/aredn-manager/internal/server/api/websocket"
	"github.com/USA-RedDragon/aredn-manager/internal/server/websocket"
	"github.com/gin-gonic/gin"
)

// ApplyRoutes to the HTTP Mux.
func ApplyRoutes(router *gin.Engine, eventsChannel chan events.Event, config *config.Config) {
	apiV1 := router.Group("/api/v1")
	v1(apiV1, config)

	arednCompat(router)

	ws := router.Group("/ws")
	ws.GET("/events", websocket.CreateHandler(websocketControllers.CreateEventsWebsocket(eventsChannel), config))
}

func arednCompat(router *gin.Engine) {
	router.GET("/cgi-bin/sysinfo.json", v1Controllers.GETSysinfo)
	router.GET("/cgi-bin/metrics", v1Controllers.GETMetrics)
	router.GET("/cgi-bin/mesh", v1Controllers.GETMesh)
	router.GET("/a/mesh", v1Controllers.GetAMesh)
}

func v1(group *gin.RouterGroup, config *config.Config) {
	group.GET("/version", v1Controllers.GETVersion)
	group.GET("/ping", v1Controllers.GETPing)

	group.POST("/notify", v1Controllers.POSTNotify)

	group.GET("/stats", v1Controllers.GETStats)
	group.GET("/loadavg", v1Controllers.GETLoadAvg)
	group.GET("/uptime", v1Controllers.GETUptime)
	group.GET("/node-ip", v1Controllers.GETNodeIP)
	group.GET("/gridsquare", v1Controllers.GETGridsqure)
	group.GET("/hostname", v1Controllers.GETHostname)

	v1Auth := group.Group("/auth")
	v1Auth.POST("/login", v1Controllers.POSTLogin)
	v1Auth.GET("/logout", v1Controllers.GETLogout)

	v1Users := group.Group("/users")
	// Paginated
	v1Users.GET("", middleware.RequireLogin(), v1Controllers.GETUsers)
	v1Users.POST("", middleware.RequireLogin(), v1Controllers.POSTUser)
	v1Users.GET("/me", middleware.RequireLogin(), v1Controllers.GETUserSelf)
	v1Users.GET("/:id", middleware.RequireLogin(), v1Controllers.GETUser)
	v1Users.PATCH("/:id", middleware.RequireLogin(), v1Controllers.PATCHUser)
	v1Users.DELETE("/:id", middleware.RequireLogin(), v1Controllers.DELETEUser)

	v1OLSR := group.Group("/olsr")
	v1OLSR.GET("/hosts", v1Controllers.GETOLSRHosts)
	v1OLSR.GET("/hosts/count", v1Controllers.GETOLSRHostsCount)
	v1OLSR.GET("/running", v1Controllers.GETOLSRRunning)

	if config.Babel.Enabled {
		v1Babel := group.Group("/babel")
		v1Babel.GET("/running", v1Controllers.GETBabelRunning)
	}

	v1Wireguard := group.Group("/wireguard")
	v1Wireguard.GET("/genkey", v1Controllers.GETWireguardGenkey)
	v1Wireguard.POST("/pubkey", v1Controllers.POSTWireguardPubkey)

	v1DNS := group.Group("/dns")
	v1DNS.GET("/running", v1Controllers.GETDNSRunning)

	v1AREDNLink := group.Group("/arednlink")
	v1AREDNLink.GET("/running", v1Controllers.GETAREDNLinkRunning)

	v1Tunnels := group.Group("/tunnels")
	// Paginated
	v1Tunnels.GET("", v1Controllers.GETTunnels)
	v1Tunnels.POST("", middleware.RequireLogin(), v1Controllers.POSTTunnel)
	v1Tunnels.GET("/wireguard/count", v1Controllers.GETWireguardTunnelsCount)
	v1Tunnels.GET("/wireguard/count/connected", v1Controllers.GETWireguardTunnelsCountConnected)
	v1Tunnels.GET("/wireguard/client/count", v1Controllers.GETWireguardClientTunnelsCount)
	v1Tunnels.GET("/wireguard/server/count", v1Controllers.GETWireguardServerTunnelsCount)
	v1Tunnels.GET("/wireguard/client/count/connected", v1Controllers.GETWireguardClientTunnelsCountConnected)
	v1Tunnels.GET("/wireguard/server/count/connected", v1Controllers.GETWireguardServerTunnelsCountConnected)
	// v1Tunnels.GET("/:id", v1Controllers.GETTunnel)
	v1Tunnels.PATCH("", middleware.RequireLogin(), v1Controllers.PATCHTunnel)
	v1Tunnels.DELETE("/:id", middleware.RequireLogin(), v1Controllers.DELETETunnel)
}
