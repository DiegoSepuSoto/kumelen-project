package postgres

import (
	"context"
	"fmt"
	"strings"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository struct{ pool *pgxpool.Pool }

func NewStudentRepository(pool *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{pool: pool}
}

const studentColumns = "s.id, s.institution_id, s.program_id, s.email, s.full_name, s.entry_year, s.status, s.created_at, s.updated_at"

func scanStudent(row pgx.Row) (*models.Student, error) {
	var s models.Student
	if err := row.Scan(&s.ID, &s.InstitutionID, &s.ProgramID, &s.Email, &s.FullName, &s.EntryYear, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, mapError(err)
	}
	return &s, nil
}

func (r *StudentRepository) Create(ctx context.Context, s *models.Student) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO students (id, institution_id, program_id, email, full_name, entry_year, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		s.ID, s.InstitutionID, s.ProgramID, s.Email, s.FullName, s.EntryYear, s.Status, s.CreatedAt, s.UpdatedAt)
	return mapError(err)
}

func (r *StudentRepository) Get(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	return scanStudent(r.pool.QueryRow(ctx, "SELECT "+studentColumns+" FROM students s WHERE s.id = $1", id))
}

// escapeLike evita que % y _ del texto buscado actúen como comodines.
func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}

func (r *StudentRepository) List(ctx context.Context, f repository.StudentFilter) ([]models.Student, error) {
	var (
		conditions []string
		args       []any
	)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if f.InstitutionID != nil {
		add("s.institution_id = $%d", *f.InstitutionID)
	}
	if f.ProgramID != nil {
		add("s.program_id = $%d", *f.ProgramID)
	}
	if f.Status != "" {
		add("s.status = $%d", string(f.Status))
	}
	if f.Level != "" {
		add("p.level = $%d", string(f.Level))
	}
	if f.Search != "" {
		pattern := escapeLike(strings.ToLower(strings.TrimSpace(f.Search)))
		args = append(args, pattern+"%", "%"+pattern+"%")
		conditions = append(conditions, fmt.Sprintf("(s.email LIKE $%d OR lower(s.full_name) LIKE $%d)", len(args)-1, len(args)))
	}

	query := "SELECT " + studentColumns + " FROM students s JOIN programs p ON p.id = s.program_id"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	args = append(args, f.Limit, f.Offset)
	query += fmt.Sprintf(" ORDER BY s.full_name, s.id LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Student{}
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

func (r *StudentRepository) Update(ctx context.Context, s *models.Student) error {
	return expectOne(r.pool.Exec(ctx, `UPDATE students SET program_id = $2, email = $3, full_name = $4, entry_year = $5, status = $6, updated_at = $7
		WHERE id = $1`, s.ID, s.ProgramID, s.Email, s.FullName, s.EntryYear, s.Status, s.UpdatedAt))
}

func (r *StudentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return expectOne(r.pool.Exec(ctx, "DELETE FROM students WHERE id = $1", id))
}

func (r *StudentRepository) LookupByEmail(ctx context.Context, email string) (*models.StudentLookup, error) {
	var l models.StudentLookup
	err := r.pool.QueryRow(ctx, `SELECT s.status, s.entry_year, p.level, i.id, i.name, p.id, p.name
		FROM students s
		JOIN programs p ON p.id = s.program_id
		JOIN institutions i ON i.id = s.institution_id
		WHERE s.email = $1`, email).Scan(&l.Status, &l.EntryYear, &l.Level, &l.InstitutionID, &l.InstitutionName, &l.ProgramID, &l.ProgramName)
	if err != nil {
		return nil, mapError(err)
	}
	return &l, nil
}
