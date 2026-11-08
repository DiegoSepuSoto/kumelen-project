package models

// Identity es la persona autenticada por Firebase. UID es opaco: no es el correo ni el RUT.
// Email solo sirve para preguntarle a Laurelin si es estudiante: no se guarda ni se registra en logs.
type Identity struct {
	UID           string
	Email         string
	EmailVerified bool
}

// StudentStanding es lo que Laurelin responde sobre un correo. IsStudent es verdadero solo para estudiantes activos.
type StudentStanding struct {
	IsStudent       bool
	Status          string
	Level           string
	InstitutionName string
	ProgramName     string
}
