package v1

import (
	"net/http"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/models"
	customMiddleware "laurelin/src/infrastructure/http/handlers/rest/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type adminRequest struct {
	Email         string      `json:"email"`
	Role          models.Role `json:"role"`
	InstitutionID *uuid.UUID  `json:"institution_id"`
}

type adminsHandler struct{ useCase *usecase.Admins }

func (h *adminsHandler) create(c echo.Context) error {
	var req adminRequest
	if !bind(c, &req) {
		return nil
	}
	admin, err := h.useCase.Create(c.Request().Context(), customMiddleware.Admin(c),
		usecase.AdminInput{Email: req.Email, Role: req.Role, InstitutionID: req.InstitutionID})
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusCreated, admin)
}

func (h *adminsHandler) list(c echo.Context) error {
	items, err := h.useCase.List(c.Request().Context(), customMiddleware.Admin(c))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, echo.Map{"items": items})
}

func (h *adminsHandler) delete(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	if err := h.useCase.Delete(c.Request().Context(), customMiddleware.Admin(c), id); err != nil {
		return respondError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// me devuelve al administrador autenticado: la app de administración lo usa para saber qué mostrar.
func me(c echo.Context) error {
	return c.JSON(http.StatusOK, customMiddleware.Admin(c))
}
