package models

// Identity es la persona autenticada por Firebase. UID es opaco: no es el correo ni el RUT.
type Identity struct {
	UID string
}
