// Package postgres implementa los repositorios sobre PostgreSQL con pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"laurelin/src/domain/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect abre el pool y comprueba que la base responde.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("can't configure the database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("can't reach the database: %w", err)
	}
	return pool, nil
}

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// mapError traduce los errores de pgx a los del dominio. Los demás se devuelven tal cual (son un 500).
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == uniqueViolation || pgErr.Code == foreignKeyViolation) {
		return repository.ErrConflict
	}
	return err
}

// expectOne convierte "ninguna fila afectada" en ErrNotFound.
func expectOne(tag pgconn.CommandTag, err error) error {
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
