package v1

import (
	"net/http"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/models"
	customMiddleware "laurelin/src/infrastructure/http/handlers/rest/middleware"

	"github.com/labstack/echo/v4"
)

type programRequest struct {
	Name  string       `json:"name"`
	Level models.Level `json:"level"`
}

type programsHandler struct{ useCase *usecase.Programs }

func (h *programsHandler) create(c echo.Context) error {
	institutionID, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	var req programRequest
	if !bind(c, &req) {
		return nil
	}
	program, err := h.useCase.Create(c.Request().Context(), customMiddleware.Admin(c), institutionID, usecase.ProgramInput{Name: req.Name, Level: req.Level})
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusCreated, program)
}

func (h *programsHandler) listByInstitution(c echo.Context) error {
	institutionID, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	items, err := h.useCase.List(c.Request().Context(), customMiddleware.Admin(c), institutionID)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, echo.Map{"items": items})
}

func (h *programsHandler) get(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	program, err := h.useCase.Get(c.Request().Context(), customMiddleware.Admin(c), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, program)
}

func (h *programsHandler) update(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	var req programRequest
	if !bind(c, &req) {
		return nil
	}
	program, err := h.useCase.Update(c.Request().Context(), customMiddleware.Admin(c), id, usecase.ProgramInput{Name: req.Name, Level: req.Level})
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, program)
}

func (h *programsHandler) delete(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	if err := h.useCase.Delete(c.Request().Context(), customMiddleware.Admin(c), id); err != nil {
		return respondError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
