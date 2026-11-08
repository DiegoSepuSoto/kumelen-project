package v1

import (
	"net/http"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"
	customMiddleware "laurelin/src/infrastructure/http/handlers/rest/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type studentRequest struct {
	ProgramID uuid.UUID            `json:"program_id"`
	Email     string               `json:"email"`
	FullName  string               `json:"full_name"`
	EntryYear int                  `json:"entry_year"`
	Status    models.StudentStatus `json:"status"`
}

func (r studentRequest) input() usecase.StudentInput {
	return usecase.StudentInput{ProgramID: r.ProgramID, Email: r.Email, FullName: r.FullName, EntryYear: r.EntryYear, Status: r.Status}
}

type studentsHandler struct{ useCase *usecase.Students }

func (h *studentsHandler) create(c echo.Context) error {
	var req studentRequest
	if !bind(c, &req) {
		return nil
	}
	student, err := h.useCase.Create(c.Request().Context(), customMiddleware.Admin(c), req.input())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusCreated, student)
}

// list acepta los filtros institution_id, program_id, level, status, q (correo o nombre), limit y offset.
func (h *studentsHandler) list(c echo.Context) error {
	filter := repository.StudentFilter{
		Level:  models.Level(c.QueryParam("level")),
		Status: models.StudentStatus(c.QueryParam("status")),
		Search: c.QueryParam("q"),
		Limit:  atoiOrZero(c.QueryParam("limit")),
		Offset: atoiOrZero(c.QueryParam("offset")),
	}
	for name, target := range map[string]**uuid.UUID{"institution_id": &filter.InstitutionID, "program_id": &filter.ProgramID} {
		if value := c.QueryParam(name); value != "" {
			id, err := uuid.Parse(value)
			if err != nil {
				return badRequest(c, "the "+name+" filter is not a valid identifier")
			}
			*target = &id
		}
	}
	items, err := h.useCase.List(c.Request().Context(), customMiddleware.Admin(c), filter)
	if err != nil {
		return respondError(c, err)
	}
	limit, offset := filter.Limit, filter.Offset
	if limit <= 0 {
		limit = usecase.DefaultPageSize
	}
	if limit > usecase.MaxPageSize {
		limit = usecase.MaxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	return c.JSON(http.StatusOK, page[models.Student]{Items: items, Limit: limit, Offset: offset})
}

func (h *studentsHandler) get(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	student, err := h.useCase.Get(c.Request().Context(), customMiddleware.Admin(c), id)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, student)
}

func (h *studentsHandler) update(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	var req studentRequest
	if !bind(c, &req) {
		return nil
	}
	student, err := h.useCase.Update(c.Request().Context(), customMiddleware.Admin(c), id, req.input())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, student)
}

func (h *studentsHandler) delete(c echo.Context) error {
	id, ok := pathID(c, "id")
	if !ok {
		return nil
	}
	if err := h.useCase.Delete(c.Request().Context(), customMiddleware.Admin(c), id); err != nil {
		return respondError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
