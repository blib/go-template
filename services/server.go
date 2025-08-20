package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Server configuration constants.
const (
	serverHostKey         = "server.host"
	serverPortKey         = "server.port"
	serverCertPathKey     = "server.cert"
	serverKeyPathKey      = "server.key"
	serverReadTimeoutKey  = "server.read_timeout"
	serverWriteTimeoutKey = "server.write_timeout"
	serverModeKey         = "server.mode"
	trustedProxiesKey     = "server.trusted_proxies"

	// CORS configuration.
	corsAllowOriginsKey     = "cors.allow_origins"
	corsAllowMethodsKey     = "cors.allow_methods"
	corsAllowHeadersKey     = "cors.allow_headers"
	corsExposeHeadersKey    = "cors.expose_headers"
	corsAllowCredentialsKey = "cors.allow_credentials" //nolint: gosec // G101 -- This is a key, not a secret.
	corsMaxAgeKey           = "cors.max_age"

	defaultServerHost     = "0.0.0.0"
	defaultServerPort     = 8443
	defaultServerCertPath = "cert.pem"
	defaultServerKeyPath  = "key.pem"
)

type HTTPServer struct {
	server *http.Server
	logger *zap.Logger
	config ConfigProvider
	routes [][]Route
}

type InParams struct {
	fx.In
	Config ConfigProvider
	Routes [][]Route `group:"routes"`
	Logger *zap.Logger
}

func NewHTTPServer(in InParams) *HTTPServer {
	config := in.Config

	// Set default configuration values
	config.SetDefault(serverHostKey, defaultServerHost)
	config.SetDefault(serverPortKey, defaultServerPort)
	config.SetDefault(serverCertPathKey, defaultServerCertPath)
	config.SetDefault(serverKeyPathKey, defaultServerKeyPath)
	config.SetDefault(serverReadTimeoutKey, "15s")
	config.SetDefault(serverWriteTimeoutKey, "15s")
	config.SetDefault(serverModeKey, "release")

	// Set CORS defaults (configurable via environment variables or config file)
	config.SetDefault(corsAllowOriginsKey, []string{"*"})
	config.SetDefault(corsAllowMethodsKey, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	config.SetDefault(corsAllowHeadersKey, []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"})
	config.SetDefault(corsExposeHeadersKey, []string{"Content-Length"})
	config.SetDefault(corsAllowCredentialsKey, false)
	config.SetDefault(corsMaxAgeKey, "12h")

	// Set trusted proxies defaults (configurable via environment variables or config file)
	config.SetDefault(trustedProxiesKey, []string{})

	return &HTTPServer{
		logger: in.Logger,
		config: in.Config,
		routes: in.Routes,
	}
}

func (s *HTTPServer) setupRouter() *gin.Engine {
	gin.SetMode(s.config.GetString(serverModeKey))

	router := gin.New()

	trustedProxies := s.config.GetStringSlice(trustedProxiesKey)
	if len(trustedProxies) == 0 {
		_ = router.SetTrustedProxies(nil)
	} else {
		_ = router.SetTrustedProxies(trustedProxies)
	}

	corsConfig := cors.Config{
		AllowOrigins:     s.config.GetStringSlice(corsAllowOriginsKey),
		AllowMethods:     s.config.GetStringSlice(corsAllowMethodsKey),
		AllowHeaders:     s.config.GetStringSlice(corsAllowHeadersKey),
		ExposeHeaders:    s.config.GetStringSlice(corsExposeHeadersKey),
		AllowCredentials: s.config.GetBool(corsAllowCredentialsKey),
		MaxAge:           s.config.GetDuration(corsMaxAgeKey),
	}

	router.Use(cors.New(corsConfig))
	router.Use(s.ginLogger(), gin.Recovery())

	for _, routes := range s.routes {
		for _, route := range routes {
			switch route.Method {
			case http.MethodGet:
				router.GET(route.Path, route.Handler)
			case http.MethodPost:
				router.POST(route.Path, route.Handler)
			case http.MethodPut:
				router.PUT(route.Path, route.Handler)
			case http.MethodDelete:
				router.DELETE(route.Path, route.Handler)
			case http.MethodPatch:
				router.PATCH(route.Path, route.Handler)
			case http.MethodOptions:
				router.OPTIONS(route.Path, route.Handler)
			default:
				panic(fmt.Sprintf("invalid method: %s for path: %s", route.Method, route.Path))
			}
		}
	}

	return router
}

func (s *HTTPServer) ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		s.logger.Info("HTTP Request",
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		)
	}
}

func (s *HTTPServer) Start() error {
	router := s.setupRouter()

	host := s.config.GetString(serverHostKey)
	port := s.config.GetInt(serverPortKey)
	certPath := s.config.GetString(serverCertPathKey)
	keyPath := s.config.GetString(serverKeyPathKey)

	address := fmt.Sprintf("%s:%d", host, port)

	s.server = &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  s.config.GetDuration(serverReadTimeoutKey),
		WriteTimeout: s.config.GetDuration(serverWriteTimeoutKey),
	}

	s.logger.Info("Starting HTTPS server",
		zap.String("address", address),
		zap.String("cert", certPath),
		zap.String("key", keyPath),
	)

	if err := s.server.ListenAndServeTLS(certPath, keyPath); err != nil && err != http.ErrServerClosed {
		s.logger.Error("Failed to start HTTPS server", zap.Error(err))
		return fmt.Errorf("failed to start HTTPS server: %w", err)
	}

	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTPS server")

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("Error during server shutdown", zap.Error(err))
			return fmt.Errorf("error during server shutdown: %w", err)
		}
	}

	s.logger.Info("HTTPS server stopped")
	return nil
}
