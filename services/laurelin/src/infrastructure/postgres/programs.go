package postgres

import (
	"context"

	"laurelin/src/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProgramRepository struct{ pool *pgxpool.Pool }

func NewProgramRepository(pool *pgxpool.Pool) *ProgramRepository {
	return &ProgramRepository{pool: pool}
}

const programColumns = "id, institution_id, name, level, created_at, updated_at"

func scanProgram(row pgx.Row) (*models.Program, error) {
	var p models.Program
	if err := row.Scan(&p.ID, &p.InstitutionID, &p.Name, &p.Level, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, mapError(err)
	}
	return &p, nil
}

func (r *ProgramRepository) Create(ctx context.Context, p *models.Program) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO programs ("+programColumns+") VALUES ($1, $2, $3, $4, $5, $6)",
		p.ID, p.InstitutionID, p.Name, p.Level, p.CreatedAt, p.UpdatedAt)
	return mapError(err)
}

func (r *ProgramRepository) Get(ctx context.Context, id uuid.UUID) (*models.Program, error) {
	return scanProgram(r.pool.QueryRow(ctx, "SELECT "+programColumns+" FROM programs WHERE id = $1", id))
}

func (r *ProgramRepository) ListByInstitution(ctx context.Context, institutionID uuid.UUID) ([]models.Program, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+programColumns+" FROM programs WHERE institution_id = $1 ORDER BY level, name", institutionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Program{}
	for rows.Next() {
		p, err := scanProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

func (r *ProgramRepository) Update(ctx context.Context, p *models.Program) error {
	return expectOne(r.pool.Exec(ctx, "UPDATE programs SET name = $2, level = $3, updated_at = $4 WHERE id = $1",
		p.ID, p.Name, p.Level, p.UpdatedAt))
}

func (r *ProgramRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return expectOne(r.pool.Exec(ctx, "DELETE FROM programs WHERE id = $1", id))
}
