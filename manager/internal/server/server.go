package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/USA-RedDragon/aredn-manager/internal/bandwidth"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"
)

const defTimeout = 10 * time.Second
const debugWriteTimeout = 60 * time.Second
const rateLimitRate = time.Second
const rateLimitLimit = 10

type Server struct {
	config          *config.Config
	server          *http.Server
	db              *gorm.DB
	shutdownChannel chan bool
	stats           *bandwidth.StatCounterManager
}

func NewServer(config *config.Config, db *gorm.DB, stats *bandwidth.StatCounterManager) *Server {
	return &Server{
		config:          config,
		db:              db,
		shutdownChannel: make(chan bool),
		stats:           stats,
	}
}

func (s *Server) Run() error {
	if s.config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	s.addMiddleware(r)

	api.ApplyRoutes(r, s.config)

	writeTimeout := defTimeout
	if s.config.Debug {
		writeTimeout = debugWriteTimeout
	}

	err := r.SetTrustedProxies(s.config.TrustedProxies)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", "0.0.0.0", s.config.Port),
		Handler:      r,
		ReadTimeout:  defTimeout,
		WriteTimeout: writeTimeout,
	}
	server.SetKeepAlivesEnabled(false)

	s.server = server

	go s.run()

	return nil
}

func (s *Server) run() {
	func() {
		err := s.server.ListenAndServe()
		if err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				s.shutdownChannel <- true
				return
			default:
				fmt.Printf("Failed to start HTTP server: %s", err)
			}
		}
	}()
}

func (s *Server) Stop() {
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to shutdown HTTP server: %s", err)
	}
	<-s.shutdownChannel
}

func (s *Server) addMiddleware(r *gin.Engine) {
	// Debug
	if s.config.Debug {
		pprof.Register(r)
	}

	// Tracing
	if s.config.OTLPEndpoint != "" {
		r.Use(otelgin.Middleware("api"))
		r.Use(middleware.TracingProvider(s.config))
	}

	// DBs
	r.Use(middleware.ConfigProvider(s.config))
	r.Use(middleware.DatabaseProvider(s.db))
	r.Use(middleware.OLSRDProvider(&olsrd.Parsers{
		HostsParser:    olsrd.NewHostsParser(),
		ServicesParser: olsrd.NewServicesParser(),
	}))
	r.Use(middleware.NetworkStats(s.stats))
	r.Use(middleware.PaginatedDatabaseProvider(s.db, middleware.PaginationConfig{}))

	// CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowOrigins = s.config.CORSHosts
	r.Use(cors.New(corsConfig))

	ratelimitStore := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  rateLimitRate,
		Limit: rateLimitLimit,
	})
	ratelimitMW := ratelimit.RateLimiter(ratelimitStore, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})

	r.Use(ratelimitMW)

	// Sessions
	sessionStore := gormsessions.NewStore(s.db, true, s.config.SessionSecret)
	r.Use(sessions.Sessions("sessions", sessionStore))
}
