package repository

import (
	"context"
	"iluvatar/src/domain/models"
)

// StudentDirectory responde si un correo es de un estudiante (lo implementa Laurelin).
type StudentDirectory interface {
	// Lookup devuelve IsStudent=false para un correo desconocido o de alguien que ya no es estudiante activo.
	// Devuelve error solo si no pudo averiguarlo (Laurelin caído, token de servicio incorrecto).
	Lookup(ctx context.Context, email string) (*models.StudentStanding, error)
}
