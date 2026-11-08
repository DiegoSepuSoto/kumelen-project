package usecase

import (
	"context"
	"time"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
)

type ProgramInput struct {
	Name  string
	Level models.Level
}

type Programs struct {
	repo         repository.ProgramRepository
	institutions repository.InstitutionRepository
	audit        audit
	now          clock
}

func NewPrograms(repo repository.ProgramRepository, institutions repository.InstitutionRepository, auditRepo repository.AuditRepository) *Programs {
	return &Programs{repo: repo, institutions: institutions, audit: audit{repo: auditRepo, now: time.Now}, now: time.Now}
}

func (u *Programs) validate(in ProgramInput) (string, error) {
	if !in.Level.Valid() {
		return "", invalid("the level must be undergraduate, master, doctorate or other_postgraduate")
	}
	return checkText("the name", in.Name, 2, 200)
}

func (u *Programs) Create(ctx context.Context, actor *models.Admin, institutionID uuid.UUID, in ProgramInput) (*models.Program, error) {
	if !actor.CanAccessInstitution(institutionID) {
		return nil, ErrForbidden
	}
	name, err := u.validate(in)
	if err != nil {
		return nil, err
	}
	if _, err := u.institutions.Get(ctx, institutionID); err != nil {
		return nil, err
	}
	now := u.now()
	program := &models.Program{ID: newID(), InstitutionID: institutionID, Name: name, Level: in.Level, CreatedAt: now, UpdatedAt: now}
	if err := u.repo.Create(ctx, program); err != nil {
		return nil, err
	}
	u.audit.record(ctx, actor, "program.create", "program", program.ID, &institutionID)
	return program, nil
}

func (u *Programs) List(ctx context.Context, actor *models.Admin, institutionID uuid.UUID) ([]models.Program, error) {
	if !actor.CanAccessInstitution(institutionID) {
		return nil, repository.ErrNotFound
	}
	return u.repo.ListByInstitution(ctx, institutionID)
}

// Get devuelve ErrNotFound también cuando el programa es de otra institución.
func (u *Programs) Get(ctx context.Context, actor *models.Admin, id uuid.UUID) (*models.Program, error) {
	program, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.CanAccessInstitution(program.InstitutionID) {
		return nil, repository.ErrNotFound
	}
	return program, nil
}

func (u *Programs) Update(ctx context.Context, actor *models.Admin, id uuid.UUID, in ProgramInput) (*models.Program, error) {
	program, err := u.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	name, err := u.validate(in)
	if err != nil {
		return nil, err
	}
	program.Name, program.Level, program.UpdatedAt = name, in.Level, u.now()
	if err := u.repo.Update(ctx, program); err != nil {
		return nil, err
	}
	u.audit.record(ctx, actor, "program.update", "program", id, &program.InstitutionID)
	return program, nil
}

// Delete falla con ErrConflict mientras el programa tenga estudiantes.
func (u *Programs) Delete(ctx context.Context, actor *models.Admin, id uuid.UUID) error {
	program, err := u.Get(ctx, actor, id)
	if err != nil {
		return err
	}
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	u.audit.record(ctx, actor, "program.delete", "program", id, &program.InstitutionID)
	return nil
}
