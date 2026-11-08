package v1

import (
	"laurelin/src/application/usecase"

	"github.com/labstack/echo/v4"
)

type Dependencies struct {
	Institutions *usecase.Institutions
	Programs     *usecase.Programs
	Students     *usecase.Students
	Admins       *usecase.Admins
}

// Register monta las rutas /v1. auth debe dejar al administrador en el contexto (middleware.RequireAdmin).
func Register(e *echo.Echo, auth echo.MiddlewareFunc, d Dependencies) {
	g := e.Group("/v1", auth)
	institutions := &institutionsHandler{useCase: d.Institutions}
	programs := &programsHandler{useCase: d.Programs}
	students := &studentsHandler{useCase: d.Students}
	admins := &adminsHandler{useCase: d.Admins}

	g.GET("/me", me)

	g.POST("/institutions", institutions.create)
	g.GET("/institutions", institutions.list)
	g.GET("/institutions/:id", institutions.get)
	g.PUT("/institutions/:id", institutions.update)
	g.DELETE("/institutions/:id", institutions.delete)

	g.POST("/institutions/:id/programs", programs.create)
	g.GET("/institutions/:id/programs", programs.listByInstitution)
	g.GET("/programs/:id", programs.get)
	g.PUT("/programs/:id", programs.update)
	g.DELETE("/programs/:id", programs.delete)

	g.POST("/students", students.create)
	g.GET("/students", students.list)
	g.GET("/students/:id", students.get)
	g.PUT("/students/:id", students.update)
	g.DELETE("/students/:id", students.delete)

	g.POST("/admins", admins.create)
	g.GET("/admins", admins.list)
	g.DELETE("/admins/:id", admins.delete)
}
