// Package repository define las interfaces que implementa la infraestructura. No importa de otras capas.
package repository

import (
	"context"
	"errors"

	"laurelin/src/domain/models"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict") // duplicado, o se intenta borrar algo que aún se usa
	ErrInvalidToken = errors.New("invalid token")
)

// TokenVerifier valida un ID token de Firebase.
type TokenVerifier interface {
	Verify(ctx context.Context, idToken string) (*models.Identity, error)
}

type InstitutionRepository interface {
	Create(ctx context.Context, institution *models.Institution) error
	Get(ctx context.Context, id uuid.UUID) (*models.Institution, error)
	// List devuelve todas, o solo la indicada cuando scope no es nil.
	List(ctx context.Context, scope *uuid.UUID) ([]models.Institution, error)
	Update(ctx context.Context, institution *models.Institution) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProgramRepository interface {
	Create(ctx context.Context, program *models.Program) error
	Get(ctx context.Context, id uuid.UUID) (*models.Program, error)
	ListByInstitution(ctx context.Context, institutionID uuid.UUID) ([]models.Program, error)
	Update(ctx context.Context, program *models.Program) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type StudentFilter struct {
	InstitutionID *uuid.UUID
	ProgramID     *uuid.UUID
	Level         models.Level
	Status        models.StudentStatus
	// Search busca por prefijo del correo o por parte del nombre, sin distinguir mayúsculas.
	Search string
	Limit  int
	Offset int
}

type StudentRepository interface {
	Create(ctx context.Context, student *models.Student) error
	Get(ctx context.Context, id uuid.UUID) (*models.Student, error)
	List(ctx context.Context, filter StudentFilter) ([]models.Student, error)
	Update(ctx context.Context, student *models.Student) error
	Delete(ctx context.Context, id uuid.UUID) error
	// LookupByEmail devuelve ErrNotFound si no hay un estudiante con ese correo (ya normalizado).
	LookupByEmail(ctx context.Context, email string) (*models.StudentLookup, error)
}

type AdminRepository interface {
	Create(ctx context.Context, admin *models.Admin) error
	GetByEmail(ctx context.Context, email string) (*models.Admin, error)
	List(ctx context.Context) ([]models.Admin, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuditRepository interface {
	Record(ctx context.Context, entry *models.AuditEntry) error
}
