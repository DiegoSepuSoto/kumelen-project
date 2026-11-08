package middleware

import (
	"github.com/labstack/echo/v4"
	"iluvatar/src/domain/repository"
	"log/slog"
	"net/http"
	"strings"
)

const (
	bearerPrefix  = "Bearer "
	uidContextKey = "uid"
)

// RequireAuth exige un ID token de Firebase en la cabecera Authorization (Bearer) y deja el uid en el contexto.
// Cualquier fallo de verificación responde 401 sin detallar el motivo al cliente.
func RequireAuth(verifier repository.TokenVerifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			idToken, hasPrefix := strings.CutPrefix(header, bearerPrefix)
			if !hasPrefix || strings.TrimSpace(idToken) == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "missing bearer token"})
			}

			identity, err := verifier.Verify(c.Request().Context(), strings.TrimSpace(idToken))
			if err != nil {
				slog.Warn("rejected identity token", "error", err.Error())
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid token"})
			}

			c.Set(uidContextKey, identity.UID)
			return next(c)
		}
	}
}

// UID devuelve el uid de Firebase de quien hizo la petición; vacío si la ruta no pasó por RequireAuth.
func UID(c echo.Context) string {
	uid, _ := c.Get(uidContextKey).(string)
	return uid
}
