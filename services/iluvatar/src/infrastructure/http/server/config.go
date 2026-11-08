// Package server configura el servidor HTTP para producción: CORS, límites, cabeceras de seguridad,
// rate limiting y logs estructurados sin datos personales.
package server

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	productionEnvironment = "production"

	defaultRateLimitPerSecond = 20
	defaultBodyLimit          = "1M"
)

type Config struct {
	// Environment es ENV; "production" activa los valores seguros por defecto.
	Environment string
	// AllowedOrigins son los orígenes web que pueden llamar a la API (CORS_ALLOWED_ORIGINS, separados por coma).
	AllowedOrigins []string
	// RateLimitPerSecond son las peticiones por segundo y por IP (RATE_LIMIT_RPS); 0 lo desactiva.
	RateLimitPerSecond float64
	// TrustProxyHeaders confía en X-Forwarded-For para obtener la IP del cliente (TRUST_PROXY_HEADERS).
	// Solo debe activarse detrás de un proxy propio: si no, cualquiera podría falsear su IP.
	TrustProxyHeaders bool
	// SwaggerEnabled publica la documentación (SWAGGER_ENABLED); por defecto solo fuera de producción.
	SwaggerEnabled bool
	BodyLimit      string
}

func ConfigFromEnv() Config {
	environment := os.Getenv("ENV")
	config := Config{
		Environment:        environment,
		AllowedOrigins:     splitList(os.Getenv("CORS_ALLOWED_ORIGINS")),
		RateLimitPerSecond: defaultRateLimitPerSecond,
		TrustProxyHeaders:  parseBool(os.Getenv("TRUST_PROXY_HEADERS"), false),
		BodyLimit:          defaultBodyLimit,
	}
	config.SwaggerEnabled = parseBool(os.Getenv("SWAGGER_ENABLED"), !config.Production())

	if value := os.Getenv("RATE_LIMIT_RPS"); value != "" {
		if rate, err := strconv.ParseFloat(value, 64); err == nil && rate >= 0 {
			config.RateLimitPerSecond = rate
		}
	}
	return config
}

func (c Config) Production() bool {
	return strings.EqualFold(c.Environment, productionEnvironment)
}

// Server timeouts: acotan conexiones lentas (slowloris) y respuestas que no terminan.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func parseBool(value string, fallback bool) bool {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}
