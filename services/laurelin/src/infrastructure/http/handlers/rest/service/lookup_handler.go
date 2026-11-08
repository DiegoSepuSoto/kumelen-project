// Package service expone rutas para otros servicios (no para personas), protegidas con un token compartido.
package service

import (
	"errors"
	"log/slog"
	"net/http"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/labstack/echo/v4"
)

type lookupRequest struct {
	Email string `json:"email"`
}

type lookupResponse struct {
	// IsStudent es true solo si el estudiante está activo (no egresado, retirado ni con suspensión).
	IsStudent bool `json:"is_student"`
	*models.StudentLookup
}

// Register monta POST /internal/v1/students/lookup. El correo va en el cuerpo y no en la URL
// para que no quede en logs de proxys. Un correo desconocido responde 404.
func Register(e *echo.Echo, serviceAuth echo.MiddlewareFunc, students *usecase.Students) {
	e.POST("/internal/v1/students/lookup", func(c echo.Context) error {
		var req lookupRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "the request body is not valid"})
		}
		lookup, err := students.Lookup(c.Request().Context(), req.Email)
		switch {
		case errors.Is(err, usecase.ErrInvalid):
			return c.JSON(http.StatusBadRequest, echo.Map{"message": usecase.Message(err)})
		case errors.Is(err, repository.ErrNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"message": "not found"})
		case err != nil:
			slog.Error("lookup failed", "error", err.Error())
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "internal error"})
		}
		return c.JSON(http.StatusOK, lookupResponse{IsStudent: lookup.Status == models.StatusActive, StudentLookup: lookup})
	}, serviceAuth)
}
