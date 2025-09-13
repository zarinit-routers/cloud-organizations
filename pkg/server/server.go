package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/zarinit-routers/cloud-organizations/pkg/metrics"

	"github.com/zarinit-routers/cloud-organizations/pkg/handlers/node"
	orgHandlers "github.com/zarinit-routers/cloud-organizations/pkg/handlers/organizations"
	"github.com/zarinit-routers/cloud-organizations/pkg/logger"
	reqid "github.com/zarinit-routers/cloud-organizations/pkg/middleware/requestid"
	contracts "github.com/zarinit-routers/cloud-organizations/pkg/services/contracts"
	"github.com/zarinit-routers/middleware/auth"
)

const (
	HeaderReqID = "X-Request-ID"
)

// Server returns a ready-to-run gin.Engine with all routes configured.
func Server(svc contracts.OrganizationsService) *gin.Engine {
	// Logger must be configured by main
	logger.Setup("organizations")

	if viper.GetString("server.mode") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(reqid.Middleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{viper.GetString("client.address")},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", HeaderReqID},
		ExposeHeaders:    []string{HeaderReqID},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Prometheus metrics middleware
	r.Use(metrics.Middleware())

	// basic access log
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
			"request_id", c.Writer.Header().Get(HeaderReqID),
		)
	})

	// Probes
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/readyz", func(c *gin.Context) { c.Status(http.StatusOK) })
	// Metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API
	api := r.Group("/api/v1/organizations")
	h := orgHandlers.NewHandlers(svc)
	{
		api.POST("/", auth.Middleware(auth.AdminOnly()), h.Create)
		api.GET("/:id", auth.Middleware(), h.Get)
		api.GET("/", auth.Middleware(), h.List)
		api.PUT("/:id", auth.Middleware(auth.AdminOnly()), h.Replace)
		api.PATCH("/:id", auth.Middleware(auth.AdminOnly()), h.Patch)
		api.DELETE("/:id", auth.Middleware(auth.AdminOnly()), h.Delete)
		api.POST("/:id/restore", auth.Middleware(auth.AdminOnly()), h.Restore)

		api.POST("/bulk", auth.Middleware(auth.AdminOnly()), h.BulkCreate)
		api.PATCH("/bulk", auth.Middleware(auth.AdminOnly()), h.BulkUpdate)
		api.DELETE("/bulk", auth.Middleware(auth.AdminOnly()), h.BulkDelete)
	}

	// Compatibility endpoint
	r.POST("/api/v1/organizations/authorize-node", node.AuthorizeNode())

	return r
}

func Addr() string { return fmt.Sprintf(":%d", viper.GetInt("server.port")) }
