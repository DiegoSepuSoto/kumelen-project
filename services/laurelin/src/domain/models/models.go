// Package models define las entidades de Laurelin: instituciones, programas, estudiantes y administradores.
// No importa de otras capas.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Level es el nivel de un programa. Pregrado es "undergraduate"; el resto es postgrado.
type Level string

const (
	LevelUndergraduate     Level = "undergraduate"
	LevelMaster            Level = "master"
	LevelDoctorate         Level = "doctorate"
	LevelOtherPostgraduate Level = "other_postgraduate" // diplomados, postítulos y similares
)

func (l Level) Valid() bool {
	switch l {
	case LevelUndergraduate, LevelMaster, LevelDoctorate, LevelOtherPostgraduate:
		return true
	}
	return false
}

func (l Level) Postgraduate() bool { return l.Valid() && l != LevelUndergraduate }

// StudentStatus es la situación de un estudiante en su programa.
type StudentStatus string

const (
	StatusActive    StudentStatus = "active"
	StatusOnLeave   StudentStatus = "on_leave" // suspensión o congelamiento
	StatusGraduated StudentStatus = "graduated"
	StatusWithdrawn StudentStatus = "withdrawn"
)

func (s StudentStatus) Valid() bool {
	switch s {
	case StatusActive, StatusOnLeave, StatusGraduated, StatusWithdrawn:
		return true
	}
	return false
}

type Institution struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"` // identificador corto y estable, por ejemplo "utem"
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Program struct {
	ID            uuid.UUID `json:"id"`
	InstitutionID uuid.UUID `json:"institution_id"`
	Name          string    `json:"name"`
	Level         Level     `json:"level"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Student es un registro mínimo a propósito: no guarda RUT, teléfono ni dirección (minimización de datos).
// Email es el correo institucional en minúsculas y es único en todo Laurelin.
type Student struct {
	ID            uuid.UUID     `json:"id"`
	InstitutionID uuid.UUID     `json:"institution_id"`
	ProgramID     uuid.UUID     `json:"program_id"`
	Email         string        `json:"email"`
	FullName      string        `json:"full_name"`
	EntryYear     int           `json:"entry_year"`
	Status        StudentStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// StudentLookup es lo que Laurelin entrega a otros servicios (Kümelen) sobre un correo.
type StudentLookup struct {
	Status          StudentStatus `json:"status"`
	EntryYear       int           `json:"entry_year"`
	Level           Level         `json:"level"`
	InstitutionID   uuid.UUID     `json:"institution_id"`
	InstitutionName string        `json:"institution_name"`
	ProgramID       uuid.UUID     `json:"program_id"`
	ProgramName     string        `json:"program_name"`
}

type Role string

const (
	RoleSuperAdmin       Role = "super_admin"       // administración central: ve y modifica todo
	RoleInstitutionAdmin Role = "institution_admin" // solo su institución
)

func (r Role) Valid() bool { return r == RoleSuperAdmin || r == RoleInstitutionAdmin }

// Admin es quien opera el registro. Se identifica por el correo verificado de su cuenta de Firebase.
type Admin struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	Role          Role       `json:"role"`
	InstitutionID *uuid.UUID `json:"institution_id,omitempty"` // solo para institution_admin
	CreatedAt     time.Time  `json:"created_at"`
}

func (a Admin) IsSuper() bool { return a.Role == RoleSuperAdmin }

// CanAccessInstitution indica si el administrador puede ver y modificar los datos de esa institución.
func (a Admin) CanAccessInstitution(id uuid.UUID) bool {
	if a.IsSuper() {
		return true
	}
	return a.InstitutionID != nil && *a.InstitutionID == id
}

// Identity es la persona autenticada por Firebase.
type Identity struct {
	UID           string
	Email         string
	EmailVerified bool
}

// AuditEntry registra quién cambió qué. Guarda solo identificadores: nunca correos ni nombres.
type AuditEntry struct {
	ID            uuid.UUID
	ActorAdminID  uuid.UUID
	Action        string // por ejemplo "student.create"
	EntityType    string
	EntityID      uuid.UUID
	InstitutionID *uuid.UUID
	At            time.Time
}
