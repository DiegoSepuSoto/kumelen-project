// Package v1 expone el CRUD de Laurelin para administradores.
package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// respondError traduce los errores del dominio a HTTP. Lo inesperado queda en el log y el cliente solo ve un 500.
func respondError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalid):
		return c.JSON(http.StatusBadRequest, echo.Map{"message": usecase.Message(err)})
	case errors.Is(err, usecase.ErrForbidden):
		return c.JSON(http.StatusForbidden, echo.Map{"message": "forbidden"})
	case errors.Is(err, repository.ErrNotFound):
		return c.JSON(http.StatusNotFound, echo.Map{"message": "not found"})
	case errors.Is(err, repository.ErrConflict):
		return c.JSON(http.StatusConflict, echo.Map{"message": "conflict: the resource already exists or is still in use"})
	}
	slog.Error("unexpected error", "route", c.Path(), "error", err.Error())
	return c.JSON(http.StatusInternalServerError, echo.Map{"message": "internal error"})
}

func badRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, echo.Map{"message": message})
}

// bind lee el cuerpo JSON; responde 400 y devuelve false si no es válido.
func bind(c echo.Context, target any) bool {
	if err := c.Bind(target); err != nil {
		_ = badRequest(c, "the request body is not valid")
		return false
	}
	return true
}

func pathID(c echo.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		_ = badRequest(c, "the "+name+" is not a valid identifier")
		return uuid.Nil, false
	}
	return id, true
}

type page[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func atoiOrZero(value string) int {
	n, _ := strconv.Atoi(value)
	return n
}
