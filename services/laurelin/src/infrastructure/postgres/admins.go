package postgres

import (
	"context"

	"laurelin/src/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct{ pool *pgxpool.Pool }

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository { return &AdminRepository{pool: pool} }

const adminColumns = "id, email, role, institution_id, created_at"

func scanAdmin(row pgx.Row) (*models.Admin, error) {
	var a models.Admin
	if err := row.Scan(&a.ID, &a.Email, &a.Role, &a.InstitutionID, &a.CreatedAt); err != nil {
		return nil, mapError(err)
	}
	return &a, nil
}

func (r *AdminRepository) Create(ctx context.Context, a *models.Admin) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO admins ("+adminColumns+") VALUES ($1, $2, $3, $4, $5)",
		a.ID, a.Email, a.Role, a.InstitutionID, a.CreatedAt)
	return mapError(err)
}

func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (*models.Admin, error) {
	return scanAdmin(r.pool.QueryRow(ctx, "SELECT "+adminColumns+" FROM admins WHERE email = $1", email))
}

func (r *AdminRepository) List(ctx context.Context) ([]models.Admin, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+adminColumns+" FROM admins ORDER BY email")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Admin{}
	for rows.Next() {
		a, err := scanAdmin(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *a)
	}
	return items, rows.Err()
}

func (r *AdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return expectOne(r.pool.Exec(ctx, "DELETE FROM admins WHERE id = $1", id))
}
