package health

import (
	"context"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const checkTimeout = 2 * time.Second

type healthHandler struct {
	checks []Check
}

// Check es una dependencia sin la cual el servicio no puede responder bien (por ejemplo, el CMS).
type Check struct {
	Name string
	Run  func(ctx context.Context) error
}

type health struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	AppName     string `json:"app_name"`
	Environment string `json:"environment"`
}

type readiness struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// NewHealthHandler registra /health (el proceso vive) y /ready (puede atender: sus dependencias responden).
// La plataforma reinicia el contenedor si /health falla y deja de enviarle tráfico si /ready falla.
func NewHealthHandler(e *echo.Echo, checks ...Check) {
	h := &healthHandler{checks: checks}

	e.GET("/health", h.HealthCheck)
	e.GET("/ready", h.Ready)
}

func (h *healthHandler) HealthCheck(c echo.Context) error {
	healthCheck := health{
		Status:      "UP",
		AppName:     "Ilúvatar",
		Version:     os.Getenv("APP_VERSION"),
		Environment: os.Getenv("ENV"),
	}

	return c.JSON(http.StatusOK, healthCheck)
}

// Ready no detalla el motivo de un fallo al cliente (podría revelar la red interna); queda en el log.
func (h *healthHandler) Ready(c echo.Context) error {
	result := readiness{Status: "UP", Checks: map[string]string{}}

	for _, check := range h.checks {
		ctx, cancel := context.WithTimeout(c.Request().Context(), checkTimeout)
		err := check.Run(ctx)
		cancel()

		if err != nil {
			slog.Warn("readiness check failed", "check", check.Name, "error", err.Error())
			result.Status = "DOWN"
			result.Checks[check.Name] = "DOWN"
			continue
		}
		result.Checks[check.Name] = "UP"
	}

	if result.Status != "UP" {
		return c.JSON(http.StatusServiceUnavailable, result)
	}
	return c.JSON(http.StatusOK, result)
}
