package v1

import (
	"net/http"

	"laurelin/src/application/usecase"
	customMiddleware "laurelin/src/infrastructure/http/handlers/rest/middleware"

	"github.com/labstack/echo/v4"
)

type institutionRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type institutionsHandler struct{ useCase *usecase.Institutions }

func (h *institutionsHandler) create(c echo.Context) error {
	var req institutionRequest
	if !bind(c, &req) {
		return nil
	}
	institution, err := h.useCase.Create(c.Request().Context(), customMiddleware.Admin(c), usecase.InstitutionInput{Code: req.Code, Name: req.Name})
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusCreated, institution)
}

func (h *institutionsHandler) list(c echo.Context) error {
	items, err := h.useCase.List(c.Request().Context(), customMiddleware.Admin(c))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, echo.Map{"items": items})
}

func (h *institutionsHandler) get(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	institution, err := h.useCase.Get(c.Request().Context(), customMiddleware.Admin(c), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, institution)
}

func (h *institutionsHandler) update(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	var req institutionRequest
	if !bind(c, &req) {
		return nil
	}
	institution, err := h.useCase.Update(c.Request().Context(), customMiddleware.Admin(c), id, usecase.InstitutionInput{Code: req.Code, Name: req.Name})
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, institution)
}

func (h *institutionsHandler) delete(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	if err := h.useCase.Delete(c.Request().Context(), customMiddleware.Admin(c), id); err != nil {
		return respondError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
