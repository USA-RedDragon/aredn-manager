package api

import (
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	v1Controllers "github.com/USA-RedDragon/aredn-manager/internal/server/api/controllers/v1"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/middleware"
	websocketControllers "github.com/USA-RedDragon/aredn-manager/internal/server/api/websocket"
	"github.com/USA-RedDragon/aredn-manager/internal/server/websocket"
	"github.com/gin-gonic/gin"
)

// ApplyRoutes to the HTTP Mux.
func ApplyRoutes(router *gin.Engine, config *config.Config) {
	apiV1 := router.Group("/api/v1")
	v1(apiV1, config)

	ws := router.Group("/ws")
	ws.GET("/hosts", middleware.RequireLogin(config), websocket.CreateHandler(websocketControllers.CreateHostsWebsocket(), config))
}

func v1(group *gin.RouterGroup, config *config.Config) {
	group.GET("/version", v1Controllers.GETVersion)
	group.GET("/ping", v1Controllers.GETPing)

	group.POST("/notify", v1Controllers.POSTNotify)

	group.GET("/stats", v1Controllers.GETStats)

	v1Auth := group.Group("/auth")
	v1Auth.POST("/login", v1Controllers.POSTLogin)
	v1Auth.GET("/logout", v1Controllers.GETLogout)

	v1Users := group.Group("/users")
	// Paginated
	v1Users.GET("", middleware.RequireLogin(config), v1Controllers.GETUsers)
	v1Users.POST("", middleware.RequireLogin(config), v1Controllers.POSTUser)
	v1Users.GET("/me", middleware.RequireLogin(config), v1Controllers.GETUserSelf)
	v1Users.GET("/:id", middleware.RequireLogin(config), v1Controllers.GETUser)
	v1Users.PATCH("/:id", middleware.RequireLogin(config), v1Controllers.PATCHUser)
	v1Users.DELETE("/:id", middleware.RequireLogin(config), v1Controllers.DELETEUser)

	v1OLSR := group.Group("/olsr")
	v1OLSR.GET("/hosts", v1Controllers.GETOLSRHosts)
	v1OLSR.GET("/running", v1Controllers.GETOLSRRunning)

	v1VTun := group.Group("/vtun")
	v1VTun.GET("/running", v1Controllers.GETVtunRunning)

	v1DNS := group.Group("/dns")
	v1DNS.GET("/running", v1Controllers.GETDNSRunning)

	v1Tunnels := group.Group("/tunnels")
	// Paginated
	v1Tunnels.GET("", v1Controllers.GETTunnels)
	v1Tunnels.POST("", middleware.RequireLogin(config), v1Controllers.POSTTunnel)
	// v1Tunnels.GET("/:id", v1Controllers.GETTunnel)
	// v1Tunnels.PATCH("/:id", middleware.RequireLogin(config), v1Controllers.PATCHTunnel)
	v1Tunnels.DELETE("/:id", middleware.RequireLogin(config), v1Controllers.DELETETunnel)

	v1Meshes := group.Group("/meshes")
	// // Paginated
	v1Meshes.GET("", v1Controllers.GETMeshes)
	v1Meshes.POST("", middleware.RequireLogin(config), v1Controllers.POSTMesh)
	// v1Meshes.GET("/:id", v1Controllers.GETMesh)
	// v1Meshes.PATCH("/:id", middleware.RequireLogin(config), v1Controllers.PATCHMesh)
	v1Meshes.DELETE("/:id", middleware.RequireLogin(config), v1Controllers.DELETEMesh)
}