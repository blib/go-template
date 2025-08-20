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
	SERVER_HOST          = "server.host"
	SERVER_PORT          = "server.port"
	SERVER_CERT_PATH     = "server.cert"
	SERVER_KEY_PATH      = "server.key"
	SERVER_READ_TIMEOUT  = "server.read_timeout"
	SERVER_WRITE_TIMEOUT = "server.write_timeout"
	SERVER_MODE          = "server.mode"
	TRUSTED_PROXIES      = "server.trusted_proxies"

	// CORS configuration.
	CORS_ALLOW_ORIGINS     = "cors.allow_origins"
	CORS_ALLOW_METHODS     = "cors.allow_methods"
	CORS_ALLOW_HEADERS     = "cors.allow_headers"
	CORS_EXPOSE_HEADERS    = "cors.expose_headers"
	CORS_ALLOW_CREDENTIALS = "cors.allow_credentials"
	CORS_MAX_AGE           = "cors.max_age"
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
	config.SetDefault(SERVER_HOST, "0.0.0.0")
	config.SetDefault(SERVER_PORT, 8443)
	config.SetDefault(SERVER_CERT_PATH, "cert.pem")
	config.SetDefault(SERVER_KEY_PATH, "key.pem")
	config.SetDefault(SERVER_READ_TIMEOUT, "15s")
	config.SetDefault(SERVER_WRITE_TIMEOUT, "15s")
	config.SetDefault(SERVER_MODE, "release")

	// Set CORS defaults (configurable via environment variables or config file)
	config.SetDefault(CORS_ALLOW_ORIGINS, []string{"*"})
	config.SetDefault(CORS_ALLOW_METHODS, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	config.SetDefault(CORS_ALLOW_HEADERS, []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"})
	config.SetDefault(CORS_EXPOSE_HEADERS, []string{"Content-Length"})
	config.SetDefault(CORS_ALLOW_CREDENTIALS, false)
	config.SetDefault(CORS_MAX_AGE, "12h")

	// Set trusted proxies defaults (configurable via environment variables or config file)
	config.SetDefault(TRUSTED_PROXIES, []string{})

	return &HTTPServer{
		logger: in.Logger,
		config: in.Config,
		routes: in.Routes,
	}
}

func (s *HTTPServer) setupRouter() *gin.Engine {
	gin.SetMode(s.config.GetString(SERVER_MODE))

	router := gin.New()

	trustedProxies := s.config.GetStringSlice(TRUSTED_PROXIES)
	if len(trustedProxies) == 0 {
		router.SetTrustedProxies(nil)
	} else {
		router.SetTrustedProxies(trustedProxies)
	}

	corsConfig := cors.Config{
		AllowOrigins:     s.config.GetStringSlice(CORS_ALLOW_ORIGINS),
		AllowMethods:     s.config.GetStringSlice(CORS_ALLOW_METHODS),
		AllowHeaders:     s.config.GetStringSlice(CORS_ALLOW_HEADERS),
		ExposeHeaders:    s.config.GetStringSlice(CORS_EXPOSE_HEADERS),
		AllowCredentials: s.config.GetBool(CORS_ALLOW_CREDENTIALS),
		MaxAge:           s.config.GetDuration(CORS_MAX_AGE),
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

	host := s.config.GetString(SERVER_HOST)
	port := s.config.GetInt(SERVER_PORT)
	certPath := s.config.GetString(SERVER_CERT_PATH)
	keyPath := s.config.GetString(SERVER_KEY_PATH)

	address := fmt.Sprintf("%s:%d", host, port)

	s.server = &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  s.config.GetDuration(SERVER_READ_TIMEOUT),
		WriteTimeout: s.config.GetDuration(SERVER_WRITE_TIMEOUT),
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
