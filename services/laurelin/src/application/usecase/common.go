// Package usecase contiene las reglas de Laurelin: quién puede hacer qué y qué datos son válidos.
// Cada caso de uso recibe al administrador que actúa; el alcance (institución) se decide aquí, no en los handlers.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
)

var (
	// ErrForbidden: el administrador existe pero no puede hacer esto.
	ErrForbidden = errors.New("forbidden")
	// ErrInvalid: los datos enviados no son válidos. Su mensaje es seguro de mostrar al cliente.
	ErrInvalid = errors.New("invalid input")
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// Message devuelve el motivo de un ErrInvalid sin el prefijo técnico.
func Message(err error) string {
	return strings.TrimPrefix(err.Error(), ErrInvalid.Error()+": ")
}

const (
	DefaultPageSize = 50
	MaxPageSize     = 200
	maxEmailLength  = 254
)

var codePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,39}$`)

// normalizeEmail deja el correo en minúsculas y rechaza lo que no sea una dirección simple (sin nombre ni comentarios).
func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > maxEmailLength {
		return "", invalid("the email is required and must have at most %d characters", maxEmailLength)
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || !strings.Contains(email[strings.LastIndex(email, "@"):], ".") {
		return "", invalid("the email is not valid")
	}
	return email, nil
}

func checkText(field, value string, min, max int) (string, error) {
	value = strings.TrimSpace(value)
	if n := utf8.RuneCountInString(value); n < min || n > max {
		return "", invalid("%s must have between %d and %d characters", field, min, max)
	}
	return value, nil
}

type clock func() time.Time

// audit registra los cambios. Si falla, el cambio ya se hizo: se deja un error en el log (sin datos personales)
// en vez de deshacerlo. Es un límite conocido (ver CLAUDE.md).
type audit struct {
	repo repository.AuditRepository
	now  clock
}

func (a audit) record(ctx context.Context, actor *models.Admin, action, entityType string, entityID uuid.UUID, institutionID *uuid.UUID) {
	entry := &models.AuditEntry{
		ID:            uuid.Must(uuid.NewV7()),
		ActorAdminID:  actor.ID,
		Action:        action,
		EntityType:    entityType,
		EntityID:      entityID,
		InstitutionID: institutionID,
		At:            a.now(),
	}
	if err := a.repo.Record(ctx, entry); err != nil {
		slog.Error("can't record the audit entry", "action", action, "entity_id", entityID.String(), "error", err.Error())
	}
}

func newID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

func clampPage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
