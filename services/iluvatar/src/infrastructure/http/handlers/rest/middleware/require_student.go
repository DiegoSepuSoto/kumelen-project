package middleware

import (
	"github.com/labstack/echo/v4"
	"iluvatar/src/domain/repository"
	"log/slog"
	"net/http"
)

// RequireStudent exige que quien hizo la petición sea un estudiante activo según Laurelin. Va después de RequireAuth.
//   - Cuenta sin correo verificado: 403 (cualquiera podría registrar la cuenta con el correo de otra persona).
//   - Correo desconocido, o de alguien que no está activo: 403.
//   - Laurelin no responde (y no hay una respuesta anterior que sirva): 503. Se falla cerrado a propósito.
//
// El correo nunca se registra en el log ni se devuelve al cliente.
func RequireStudent(directory repository.StudentDirectory) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity := Identity(c)
			if identity == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "missing identity"})
			}
			if !identity.EmailVerified || identity.Email == "" {
				return c.JSON(http.StatusForbidden, echo.Map{"message": "a verified email is required"})
			}

			standing, err := directory.Lookup(c.Request().Context(), identity.Email)
			if err != nil {
				slog.Error("can't verify the student", "error", err.Error())
				return c.JSON(http.StatusServiceUnavailable, echo.Map{"message": "student verification unavailable"})
			}
			if !standing.IsStudent {
				return c.JSON(http.StatusForbidden, echo.Map{"message": "not an active student"})
			}
			return next(c)
		}
	}
}
