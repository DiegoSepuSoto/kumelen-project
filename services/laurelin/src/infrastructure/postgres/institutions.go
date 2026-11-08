package postgres

import (
	"context"

	"laurelin/src/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InstitutionRepository struct{ pool *pgxpool.Pool }

func NewInstitutionRepository(pool *pgxpool.Pool) *InstitutionRepository {
	return &InstitutionRepository{pool: pool}
}

const institutionColumns = "id, code, name, created_at, updated_at"

func scanInstitution(row pgx.Row) (*models.Institution, error) {
	var i models.Institution
	if err := row.Scan(&i.ID, &i.Code, &i.Name, &i.CreatedAt, &i.UpdatedAt); err != nil {
		return nil, mapError(err)
	}
	return &i, nil
}

func (r *InstitutionRepository) Create(ctx context.Context, i *models.Institution) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO institutions ("+institutionColumns+") VALUES ($1, $2, $3, $4, $5)",
		i.ID, i.Code, i.Name, i.CreatedAt, i.UpdatedAt)
	return mapError(err)
}

func (r *InstitutionRepository) Get(ctx context.Context, id uuid.UUID) (*models.Institution, error) {
	return scanInstitution(r.pool.QueryRow(ctx, "SELECT "+institutionColumns+" FROM institutions WHERE id = $1", id))
}

func (r *InstitutionRepository) List(ctx context.Context, scope *uuid.UUID) ([]models.Institution, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+institutionColumns+" FROM institutions WHERE ($1::uuid IS NULL OR id = $1) ORDER BY name", scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Institution{}
	for rows.Next() {
		i, err := scanInstitution(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *i)
	}
	return items, rows.Err()
}

func (r *InstitutionRepository) Update(ctx context.Context, i *models.Institution) error {
	return expectOne(r.pool.Exec(ctx, "UPDATE institutions SET code = $2, name = $3, updated_at = $4 WHERE id = $1",
		i.ID, i.Code, i.Name, i.UpdatedAt))
}

func (r *InstitutionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return expectOne(r.pool.Exec(ctx, "DELETE FROM institutions WHERE id = $1", id))
}
