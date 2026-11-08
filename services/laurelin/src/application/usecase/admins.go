package usecase

import (
	"context"
	"errors"
	"time"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
)

type AdminInput struct {
	Email         string
	Role          models.Role
	InstitutionID *uuid.UUID
}

type Admins struct {
	repo         repository.AdminRepository
	institutions repository.InstitutionRepository
	audit        audit
	now          clock
}

func NewAdmins(repo repository.AdminRepository, institutions repository.InstitutionRepository, auditRepo repository.AuditRepository) *Admins {
	return &Admins{repo: repo, institutions: institutions, audit: audit{repo: auditRepo, now: time.Now}, now: time.Now}
}

// Authenticate convierte la identidad de Firebase en un administrador. Exige un correo verificado:
// sin eso cualquiera podría registrar la cuenta con el correo de un administrador.
// Cualquier fallo es ErrForbidden, sin distinguir "no es administrador" de "correo sin verificar".
func (u *Admins) Authenticate(ctx context.Context, identity *models.Identity) (*models.Admin, error) {
	if identity == nil || !identity.EmailVerified {
		return nil, ErrForbidden
	}
	email, err := normalizeEmail(identity.Email)
	if err != nil {
		return nil, ErrForbidden
	}
	admin, err := u.repo.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrForbidden
	}
	return admin, err
}

// Bootstrap crea la primera cuenta de administración central si ese correo aún no existe.
// Es idempotente: se puede llamar en cada arranque.
func (u *Admins) Bootstrap(ctx context.Context, email string) error {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	_, err = u.repo.GetByEmail(ctx, normalized)
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	admin := &models.Admin{ID: newID(), Email: normalized, Role: models.RoleSuperAdmin, CreatedAt: u.now()}
	if err := u.repo.Create(ctx, admin); err != nil && !errors.Is(err, repository.ErrConflict) {
		return err
	}
	return nil
}

// Create solo la hace la administración central.
func (u *Admins) Create(ctx context.Context, actor *models.Admin, in AdminInput) (*models.Admin, error) {
	if !actor.IsSuper() {
		return nil, ErrForbidden
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return nil, err
	}
	admin := &models.Admin{ID: newID(), Email: email, Role: in.Role, CreatedAt: u.now()}
	switch in.Role {
	case models.RoleSuperAdmin:
		if in.InstitutionID != nil {
			return nil, invalid("a super_admin doesn't belong to an institution")
		}
	case models.RoleInstitutionAdmin:
		if in.InstitutionID == nil {
			return nil, invalid("an institution_admin needs an institution_id")
		}
		if _, err := u.institutions.Get(ctx, *in.InstitutionID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, invalid("the institution doesn't exist")
			}
			return nil, err
		}
		admin.InstitutionID = in.InstitutionID
	default:
		return nil, invalid("the role must be super_admin or institution_admin")
	}
	if err := u.repo.Create(ctx, admin); err != nil {
		return nil, err
	}
	u.audit.record(ctx, actor, "admin.create", "admin", admin.ID, admin.InstitutionID)
	return admin, nil
}

func (u *Admins) List(ctx context.Context, actor *models.Admin) ([]models.Admin, error) {
	if !actor.IsSuper() {
		return nil, ErrForbidden
	}
	return u.repo.List(ctx)
}

// Delete no permite que un administrador se elimine a sí mismo (podría dejar el registro sin nadie que lo administre).
func (u *Admins) Delete(ctx context.Context, actor *models.Admin, id uuid.UUID) error {
	if !actor.IsSuper() {
		return ErrForbidden
	}
	if id == actor.ID {
		return invalid("you can't delete your own account")
	}
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	u.audit.record(ctx, actor, "admin.delete", "admin", id, nil)
	return nil
}
