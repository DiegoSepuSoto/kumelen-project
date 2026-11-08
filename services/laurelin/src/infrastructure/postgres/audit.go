package postgres

import (
	"context"

	"laurelin/src/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct{ pool *pgxpool.Pool }

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository { return &AuditRepository{pool: pool} }

func (r *AuditRepository) Record(ctx context.Context, e *models.AuditEntry) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO audit_log (id, actor_admin_id, action, entity_type, entity_id, institution_id, at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, e.ID, e.ActorAdminID, e.Action, e.EntityType, e.EntityID, e.InstitutionID, e.At)
	return err
}
