package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// probePaths no cuentan para el rate limit ni ensucian los logs: los consulta la plataforma cada pocos segundos.
var probePaths = map[string]bool{"/health": true, "/ready": true}

const apiContentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'"

// Apply registra los middlewares comunes y los timeouts del servidor, en este orden:
// request id, logs, recuperación de pánicos, cabeceras de seguridad, CORS, límite de cuerpo y rate limit.
func Apply(e *echo.Echo, config Config, logger *slog.Logger) {
	// Sin esto Echo confiaría en X-Forwarded-For de cualquier cliente y el rate limit se podría evadir.
	if config.TrustProxyHeaders {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
		e.SchemeExtractor = echo.ExtractSchemeFromHeaders()
	} else {
		e.IPExtractor = echo.ExtractIPDirect()
		e.SchemeExtractor = echo.ExtractSchemeDirect()
	}

	e.Server.ReadHeaderTimeout = readHeaderTimeout
	e.Server.ReadTimeout = readTimeout
	e.Server.WriteTimeout = writeTimeout
	e.Server.IdleTimeout = idleTimeout

	e.Use(middleware.RequestID())
	e.Use(requestLogger(logger))
	e.Use(recoverPanics(logger))
	e.Use(secureHeaders(config))
	if cors := corsMiddleware(config); cors != nil {
		e.Use(cors)
	}
	e.Use(middleware.BodyLimit(config.BodyLimit))
	if config.RateLimitPerSecond > 0 {
		e.Use(rateLimiter(config.RateLimitPerSecond))
	}
}

// requestLogger registra una línea JSON por petición. Usa la ruta registrada (/v1/post/:id) y nunca la URL
// completa ni la cabecera Authorization: la URL puede traer términos de búsqueda de una persona.
func requestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper:      func(c echo.Context) bool { return probePaths[c.Path()] },
		LogMethod:    true,
		LogRoutePath: true,
		LogStatus:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogError:     true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			route := v.RoutePath
			if route == "" {
				route = "unmatched"
			}
			attributes := []any{
				"method", v.Method,
				"route", route,
				"status", v.Status,
				"latency_ms", v.Latency.Milliseconds(),
				"request_id", v.RequestID,
			}
			level := slog.LevelInfo
			switch {
			case v.Status >= http.StatusInternalServerError:
				level = slog.LevelError
			case v.Status >= http.StatusBadRequest:
				level = slog.LevelWarn
			}
			if v.Error != nil && v.Status >= http.StatusInternalServerError {
				attributes = append(attributes, "error", v.Error.Error())
			}
			logger.Log(c.Request().Context(), level, "request", attributes...)
			return nil
		},
	})
}

func secureHeaders(config Config) echo.MiddlewareFunc {
	secure := middleware.SecureConfig{
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		ReferrerPolicy:     "no-referrer",
	}
	if config.Production() {
		// Echo solo envía HSTS en peticiones HTTPS o con X-Forwarded-Proto=https.
		secure.HSTSMaxAge = 31536000
	}
	// La política restrictiva rompería la interfaz de Swagger, que necesita scripts.
	if !config.SwaggerEnabled {
		secure.ContentSecurityPolicy = apiContentSecurityPolicy
	}
	return middleware.SecureWithConfig(secure)
}

// corsMiddleware devuelve nil cuando no debe haber CORS: en producción sin orígenes configurados no se
// permite ningún origen web. Fuera de producción, sin configuración, se permite cualquiera para desarrollar.
func corsMiddleware(config Config) echo.MiddlewareFunc {
	if len(config.AllowedOrigins) == 0 {
		if config.Production() {
			return nil
		}
		return middleware.CORS()
	}
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType},
		MaxAge:       int((time.Hour).Seconds()),
	})
}

func rateLimiter(requestsPerSecond float64) echo.MiddlewareFunc {
	burst := int(requestsPerSecond * 2)
	if burst < 1 {
		burst = 1
	}
	store := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      rate.Limit(requestsPerSecond),
		Burst:     burst,
		ExpiresIn: 3 * time.Minute,
	})
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper:             func(c echo.Context) bool { return probePaths[c.Path()] || strings.HasPrefix(c.Path(), "/swagger") },
		Store:               store,
		IdentifierExtractor: func(c echo.Context) (string, error) { return c.RealIP(), nil },
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return c.JSON(http.StatusTooManyRequests, echo.Map{"message": "too many requests"})
		},
		ErrorHandler: func(c echo.Context, _ error) error {
			return c.JSON(http.StatusForbidden, echo.Map{"message": "can't identify the client"})
		},
	})
}

// recoverPanics convierte un pánico en un 500 y lo registra, con su traza, en el mismo log estructurado.
func recoverPanics(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack: true,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("panic recovered", "route", c.Path(), "error", err.Error(), "stack", string(stack))
			return err
		},
	})
}
