package usecase

import (
	"context"
	"strings"
	"time"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
)

type InstitutionInput struct {
	Code string
	Name string
}

type Institutions struct {
	repo  repository.InstitutionRepository
	audit audit
	now   clock
}

func NewInstitutions(repo repository.InstitutionRepository, auditRepo repository.AuditRepository) *Institutions {
	return &Institutions{repo: repo, audit: audit{repo: auditRepo, now: time.Now}, now: time.Now}
}

func (u *Institutions) validate(in InstitutionInput) (code, name string, err error) {
	code = strings.ToLower(strings.TrimSpace(in.Code))
	if !codePattern.MatchString(code) {
		return "", "", invalid("the code must have 2 to 40 characters: lowercase letters, digits and hyphens")
	}
	name, err = checkText("the name", in.Name, 2, 200)
	return code, name, err
}

// Create solo la hace la administración central.
func (u *Institutions) Create(ctx context.Context, actor *models.Admin, in InstitutionInput) (*models.Institution, error) {
	if !actor.IsSuper() {
		return nil, ErrForbidden
	}
	code, name, err := u.validate(in)
	if err != nil {
		return nil, err
	}
	now := u.now()
	institution := &models.Institution{ID: newID(), Code: code, Name: name, CreatedAt: now, UpdatedAt: now}
	if err := u.repo.Create(ctx, institution); err != nil {
		return nil, err
	}
	u.audit.record(ctx, actor, "institution.create", "institution", institution.ID, &institution.ID)
	return institution, nil
}

// Get devuelve ErrNotFound también cuando la institución existe pero el administrador no tiene acceso.
func (u *Institutions) Get(ctx context.Context, actor *models.Admin, id uuid.UUID) (*models.Institution, error) {
	if !actor.CanAccessInstitution(id) {
		return nil, repository.ErrNotFound
	}
	return u.repo.Get(ctx, id)
}

func (u *Institutions) List(ctx context.Context, actor *models.Admin) ([]models.Institution, error) {
	if actor.IsSuper() {
		return u.repo.List(ctx, nil)
	}
	return u.repo.List(ctx, actor.InstitutionID)
}

func (u *Institutions) Update(ctx context.Context, actor *models.Admin, id uuid.UUID, in InstitutionInput) (*models.Institution, error) {
	if !actor.IsSuper() {
		return nil, ErrForbidden
	}
	code, name, err := u.validate(in)
	if err != nil {
		return nil, err
	}
	institution, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	institution.Code, institution.Name, institution.UpdatedAt = code, name, u.now()
	if err := u.repo.Update(ctx, institution); err != nil {
		return nil, err
	}
	u.audit.record(ctx, actor, "institution.update", "institution", id, &id)
	return institution, nil
}

// Delete falla con ErrConflict mientras la institución tenga programas, estudiantes o administradores.
func (u *Institutions) Delete(ctx context.Context, actor *models.Admin, id uuid.UUID) error {
	if !actor.IsSuper() {
		return ErrForbidden
	}
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	u.audit.record(ctx, actor, "institution.delete", "institution", id, &id)
	return nil
}
