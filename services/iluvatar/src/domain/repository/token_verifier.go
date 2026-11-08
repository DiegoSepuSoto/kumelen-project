package repository

import (
	"context"
	"errors"
	"iluvatar/src/domain/models"
)

// ErrInvalidToken indica que el token de identidad es inválido, está vencido o fue revocado.
var ErrInvalidToken = errors.New("invalid identity token")

type TokenVerifier interface {
	// Verify valida el ID token de Firebase y devuelve la identidad que lo emitió.
	Verify(ctx context.Context, idToken string) (*models.Identity, error)
}
