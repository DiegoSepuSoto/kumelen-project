package usecase

import (
	"context"
	"errors"
	"time"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
)

type StudentInput struct {
	ProgramID uuid.UUID
	Email     string
	FullName  string
	EntryYear int
	Status    models.StudentStatus // vacío al crear equivale a "active"
}

type Students struct {
	repo     repository.StudentRepository
	programs repository.ProgramRepository
	audit    audit
	now      clock
}

func NewStudents(repo repository.StudentRepository, programs repository.ProgramRepository, auditRepo repository.AuditRepository) *Students {
	return &Students{repo: repo, programs: programs, audit: audit{repo: auditRepo, now: time.Now}, now: time.Now}
}

// program carga el programa del estudiante: debe existir. La institución del estudiante es la de su programa.
func (u *Students) program(ctx context.Context, id uuid.UUID) (*models.Program, error) {
	program, err := u.programs.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, invalid("the program doesn't exist")
	}
	return program, err
}

func (u *Students) fill(student *models.Student, in StudentInput) error {
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return err
	}
	name, err := checkText("the full name", in.FullName, 2, 200)
	if err != nil {
		return err
	}
	if in.EntryYear < 1900 || in.EntryYear > u.now().Year()+1 {
		return invalid("the entry year is out of range")
	}
	status := in.Status
	if status == "" {
		status = models.StatusActive
	}
	if !status.Valid() {
		return invalid("the status must be active, on_leave, graduated or withdrawn")
	}
	student.Email, student.FullName, student.EntryYear, student.Status = email, name, in.EntryYear, status
	return nil
}

func (u *Students) Create(ctx context.Context, actor *models.Admin, in StudentInput) (*models.Student, error) {
	program, err := u.program(ctx, in.ProgramID)
	if err != nil {
		return nil, err
	}
	if !actor.CanAccessInstitution(program.InstitutionID) {
		return nil, ErrForbidden
	}
	now := u.now()
	student := &models.Student{ID: newID(), InstitutionID: program.InstitutionID, ProgramID: program.ID, CreatedAt: now, UpdatedAt: now}
	if err := u.fill(student, in); err != nil {
		return nil, err
	}
	if err := u.repo.Create(ctx, student); err != nil {
		return nil, err
	}
	u.audit.record(ctx, actor, "student.create", "student", student.ID, &student.InstitutionID)
	return student, nil
}

// Get devuelve ErrNotFound también cuando el estudiante es de otra institución.
func (u *Students) Get(ctx context.Context, actor *models.Admin, id uuid.UUID) (*models.Student, error) {
	student, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.CanAccessInstitution(student.InstitutionID) {
		return nil, repository.ErrNotFound
	}
	return student, nil
}

// Update reemplaza los datos del estudiante. Puede cambiar de programa, pero no de institución.
func (u *Students) Update(ctx context.Context, actor *models.Admin, id uuid.UUID, in StudentInput) (*models.Student, error) {
	student, err := u.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	program, err := u.program(ctx, in.ProgramID)
	if err != nil {
		return nil, err
	}
	if program.InstitutionID != student.InstitutionID {
		return nil, invalid("a student can't move to a program of another institution")
	}
	student.ProgramID, student.UpdatedAt = program.ID, u.now()
	if err := u.fill(student, in); err != nil {
		return nil, err
	}
	if err := u.repo.Update(ctx, student); err != nil {
		return nil, err
	}
	u.audit.record(ctx, actor, "student.update", "student", id, &student.InstitutionID)
	return student, nil
}

func (u *Students) Delete(ctx context.Context, actor *models.Admin, id uuid.UUID) error {
	student, err := u.Get(ctx, actor, id)
	if err != nil {
		return err
	}
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	u.audit.record(ctx, actor, "student.delete", "student", id, &student.InstitutionID)
	return nil
}

// List filtra por el alcance del administrador: quien administra una institución solo ve esa.
func (u *Students) List(ctx context.Context, actor *models.Admin, filter repository.StudentFilter) ([]models.Student, error) {
	if filter.Level != "" && !filter.Level.Valid() {
		return nil, invalid("the level filter is not valid")
	}
	if filter.Status != "" && !filter.Status.Valid() {
		return nil, invalid("the status filter is not valid")
	}
	if !actor.IsSuper() {
		filter.InstitutionID = actor.InstitutionID
	}
	filter.Limit, filter.Offset = clampPage(filter.Limit, filter.Offset)
	return u.repo.List(ctx, filter)
}

// Lookup es la consulta de otros servicios: dado un correo, qué estudiante es. ErrNotFound si no existe.
func (u *Students) Lookup(ctx context.Context, email string) (*models.StudentLookup, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	return u.repo.LookupByEmail(ctx, normalized)
}
