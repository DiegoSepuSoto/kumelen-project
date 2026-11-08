// Package middleware contiene la autenticación: administradores (ID token de Firebase) y servicios (token compartido).
package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/labstack/echo/v4"
)

const (
	bearerPrefix    = "Bearer "
	adminContextKey = "admin"
)

// AdminResolver convierte la identidad de Firebase en un administrador registrado.
type AdminResolver interface {
	Authenticate(ctx context.Context, identity *models.Identity) (*models.Admin, error)
}

func bearer(c echo.Context) (string, bool) {
	token, ok := strings.CutPrefix(c.Request().Header.Get(echo.HeaderAuthorization), bearerPrefix)
	token = strings.TrimSpace(token)
	return token, ok && token != ""
}

// RequireAdmin exige un ID token de Firebase válido cuyo correo, verificado, sea el de un administrador registrado.
// Token ausente o inválido es 401; una cuenta válida que no administra nada es 403. Ningún mensaje detalla el motivo.
func RequireAdmin(verifier repository.TokenVerifier, admins AdminResolver) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			idToken, ok := bearer(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "missing bearer token"})
			}
			identity, err := verifier.Verify(c.Request().Context(), idToken)
			if err != nil {
				slog.Warn("rejected identity token", "error", err.Error())
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid token"})
			}
			admin, err := admins.Authenticate(c.Request().Context(), identity)
			if errors.Is(err, usecase.ErrForbidden) {
				return c.JSON(http.StatusForbidden, echo.Map{"message": "forbidden"})
			}
			if err != nil {
				slog.Error("can't resolve the administrator", "error", err.Error())
				return c.JSON(http.StatusInternalServerError, echo.Map{"message": "internal error"})
			}
			c.Set(adminContextKey, admin)
			return next(c)
		}
	}
}

// Admin devuelve al administrador autenticado; nil si la ruta no pasó por RequireAdmin.
func Admin(c echo.Context) *models.Admin {
	admin, _ := c.Get(adminContextKey).(*models.Admin)
	return admin
}

// RequireServiceToken protege las rutas internas (otros servicios, como iluvatar) con un token compartido.
// Si el servicio no tiene token configurado, las rutas responden 503 en vez de quedar abiertas.
func RequireServiceToken(expected string) echo.MiddlewareFunc {
	expectedHash := sha256.Sum256([]byte(expected))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if expected == "" {
				return c.JSON(http.StatusServiceUnavailable, echo.Map{"message": "service token not configured"})
			}
			token, ok := bearer(c)
			tokenHash := sha256.Sum256([]byte(token))
			// Se comparan los hashes en tiempo constante para no revelar el largo ni el prefijo del token.
			if !ok || subtle.ConstantTimeCompare(tokenHash[:], expectedHash[:]) != 1 {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid service token"})
			}
			return next(c)
		}
	}
}
