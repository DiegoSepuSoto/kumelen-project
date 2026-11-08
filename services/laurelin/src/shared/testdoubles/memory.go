// Package testdoubles tiene dobles de prueba a mano. Memory reemplaza a PostgreSQL con las mismas reglas que
// importan para el dominio (correo y código únicos, no borrar lo que aún se usa); la base real se prueba aparte
// con la etiqueta "integration".
package testdoubles

import (
	"context"
	"sort"
	"strings"
	"sync"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
)

type Memory struct {
	mu           sync.Mutex
	institutions map[uuid.UUID]models.Institution
	programs     map[uuid.UUID]models.Program
	students     map[uuid.UUID]models.Student
	admins       map[uuid.UUID]models.Admin
	Audit        []models.AuditEntry
}

func NewMemory() *Memory {
	return &Memory{
		institutions: map[uuid.UUID]models.Institution{},
		programs:     map[uuid.UUID]models.Program{},
		students:     map[uuid.UUID]models.Student{},
		admins:       map[uuid.UUID]models.Admin{},
	}
}

func (m *Memory) Institutions() repository.InstitutionRepository { return institutions{m} }
func (m *Memory) Programs() repository.ProgramRepository         { return programs{m} }
func (m *Memory) Students() repository.StudentRepository         { return students{m} }
func (m *Memory) Admins() repository.AdminRepository             { return admins{m} }
func (m *Memory) AuditLog() repository.AuditRepository           { return audit{m} }

type institutions struct{ m *Memory }

func (r institutions) Create(_ context.Context, i *models.Institution) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, other := range r.m.institutions {
		if other.Code == i.Code {
			return repository.ErrConflict
		}
	}
	r.m.institutions[i.ID] = *i
	return nil
}

func (r institutions) Get(_ context.Context, id uuid.UUID) (*models.Institution, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	i, ok := r.m.institutions[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &i, nil
}

func (r institutions) List(_ context.Context, scope *uuid.UUID) ([]models.Institution, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	items := []models.Institution{}
	for id, i := range r.m.institutions {
		if scope == nil || *scope == id {
			items = append(items, i)
		}
	}
	sort.Slice(items, func(a, b int) bool { return items[a].Name < items[b].Name })
	return items, nil
}

func (r institutions) Update(_ context.Context, i *models.Institution) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if _, ok := r.m.institutions[i.ID]; !ok {
		return repository.ErrNotFound
	}
	for id, other := range r.m.institutions {
		if id != i.ID && other.Code == i.Code {
			return repository.ErrConflict
		}
	}
	r.m.institutions[i.ID] = *i
	return nil
}

func (r institutions) Delete(_ context.Context, id uuid.UUID) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if _, ok := r.m.institutions[id]; !ok {
		return repository.ErrNotFound
	}
	for _, p := range r.m.programs {
		if p.InstitutionID == id {
			return repository.ErrConflict
		}
	}
	for _, s := range r.m.students {
		if s.InstitutionID == id {
			return repository.ErrConflict
		}
	}
	for _, a := range r.m.admins {
		if a.InstitutionID != nil && *a.InstitutionID == id {
			return repository.ErrConflict
		}
	}
	delete(r.m.institutions, id)
	return nil
}

type programs struct{ m *Memory }

func (r programs) Create(_ context.Context, p *models.Program) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, other := range r.m.programs {
		if other.InstitutionID == p.InstitutionID && other.Name == p.Name && other.Level == p.Level {
			return repository.ErrConflict
		}
	}
	r.m.programs[p.ID] = *p
	return nil
}

func (r programs) Get(_ context.Context, id uuid.UUID) (*models.Program, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	p, ok := r.m.programs[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &p, nil
}

func (r programs) ListByInstitution(_ context.Context, institutionID uuid.UUID) ([]models.Program, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	items := []models.Program{}
	for _, p := range r.m.programs {
		if p.InstitutionID == institutionID {
			items = append(items, p)
		}
	}
	sort.Slice(items, func(a, b int) bool { return items[a].Name < items[b].Name })
	return items, nil
}

func (r programs) Update(_ context.Context, p *models.Program) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if _, ok := r.m.programs[p.ID]; !ok {
		return repository.ErrNotFound
	}
	r.m.programs[p.ID] = *p
	return nil
}

func (r programs) Delete(_ context.Context, id uuid.UUID) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if _, ok := r.m.programs[id]; !ok {
		return repository.ErrNotFound
	}
	for _, s := range r.m.students {
		if s.ProgramID == id {
			return repository.ErrConflict
		}
	}
	delete(r.m.programs, id)
	return nil
}

type students struct{ m *Memory }

func (r students) Create(_ context.Context, s *models.Student) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, other := range r.m.students {
		if other.Email == s.Email {
			return repository.ErrConflict
		}
	}
	r.m.students[s.ID] = *s
	return nil
}

func (r students) Get(_ context.Context, id uuid.UUID) (*models.Student, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	s, ok := r.m.students[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &s, nil
}

func (r students) List(_ context.Context, f repository.StudentFilter) ([]models.Student, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	search := strings.ToLower(strings.TrimSpace(f.Search))
	items := []models.Student{}
	for _, s := range r.m.students {
		switch {
		case f.InstitutionID != nil && s.InstitutionID != *f.InstitutionID:
		case f.ProgramID != nil && s.ProgramID != *f.ProgramID:
		case f.Status != "" && s.Status != f.Status:
		case f.Level != "" && r.m.programs[s.ProgramID].Level != f.Level:
		case search != "" && !strings.HasPrefix(s.Email, search) && !strings.Contains(strings.ToLower(s.FullName), search):
		default:
			items = append(items, s)
		}
	}
	sort.Slice(items, func(a, b int) bool { return items[a].FullName < items[b].FullName })
	if f.Offset >= len(items) {
		return []models.Student{}, nil
	}
	items = items[f.Offset:]
	if f.Limit > 0 && len(items) > f.Limit {
		items = items[:f.Limit]
	}
	return items, nil
}

func (r students) Update(_ context.Context, s *models.Student) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if _, ok := r.m.students[s.ID]; !ok {
		return repository.ErrNotFound
	}
	for id, other := range r.m.students {
		if id != s.ID && other.Email == s.Email {
			return repository.ErrConflict
		}
	}
	r.m.students[s.ID] = *s
	return nil
}

func (r students) Delete(_ context.Context, id uuid.UUID) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if _, ok := r.m.students[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.m.students, id)
	return nil
}

func (r students) LookupByEmail(_ context.Context, email string) (*models.StudentLookup, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, s := range r.m.students {
		if s.Email == email {
			program, institution := r.m.programs[s.ProgramID], r.m.institutions[s.InstitutionID]
			return &models.StudentLookup{
				Status: s.Status, EntryYear: s.EntryYear, Level: program.Level,
				InstitutionID: institution.ID, InstitutionName: institution.Name,
				ProgramID: program.ID, ProgramName: program.Name,
			}, nil
		}
	}
	return nil, repository.ErrNotFound
}

type admins struct{ m *Memory }

func (r admins) Create(_ context.Context, a *models.Admin) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, other := range r.m.admins {
		if other.Email == a.Email {
			return repository.ErrConflict
		}
	}
	r.m.admins[a.ID] = *a
	return nil
}

func (r admins) GetByEmail(_ context.Context, email string) (*models.Admin, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, a := range r.m.admins {
		if a.Email == email {
			return &a, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (r admins) List(_ context.Context) ([]models.Admin, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	items := []models.Admin{}
	for _, a := range r.m.admins {
		items = append(items, a)
	}
	sort.Slice(items, func(a, b int) bool { return items[a].Email < items[b].Email })
	return items, nil
}

func (r admins) Delete(_ context.Context, id uuid.UUID) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if _, ok := r.m.admins[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.m.admins, id)
	return nil
}

type audit struct{ m *Memory }

func (r audit) Record(_ context.Context, e *models.AuditEntry) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.Audit = append(r.m.Audit, *e)
	return nil
}

// Verifier es un TokenVerifier de prueba: cada token conocido corresponde a una identidad.
type Verifier map[string]models.Identity

func (v Verifier) Verify(_ context.Context, idToken string) (*models.Identity, error) {
	identity, ok := v[idToken]
	if !ok {
		return nil, repository.ErrInvalidToken
	}
	return &identity, nil
}
