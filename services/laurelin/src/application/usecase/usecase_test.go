package usecase

import (
	"context"
	"errors"
	"testing"

	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"
	"laurelin/src/shared/testdoubles"

	"github.com/google/uuid"
)

type fixture struct {
	mem          *testdoubles.Memory
	institutions *Institutions
	programs     *Programs
	students     *Students
	admins       *Admins

	super          *models.Admin
	utem, uchile   *models.Institution
	utemAdmin      *models.Admin
	utemProgram    *models.Program
	uchileProgram  *models.Program
	masterProgram  *models.Program
	uchileStudents []*models.Student
}

// must devuelve el valor o hace fallar la prueba con un pánico (los datos de preparación no deben fallar).
func must[T any](value T, err error) T {
	if err != nil {
		panic("unexpected error in test setup: " + err.Error())
	}
	return value
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	mem := testdoubles.NewMemory()
	f := &fixture{
		mem:          mem,
		institutions: NewInstitutions(mem.Institutions(), mem.AuditLog()),
		programs:     NewPrograms(mem.Programs(), mem.Institutions(), mem.AuditLog()),
		students:     NewStudents(mem.Students(), mem.Programs(), mem.AuditLog()),
		admins:       NewAdmins(mem.Admins(), mem.Institutions(), mem.AuditLog()),
		super:        &models.Admin{ID: uuid.New(), Email: "root@laurelin.test", Role: models.RoleSuperAdmin},
	}
	f.utem = must(f.institutions.Create(ctx, f.super, InstitutionInput{Code: "utem", Name: "Universidad Tecnológica Metropolitana"}))
	f.uchile = must(f.institutions.Create(ctx, f.super, InstitutionInput{Code: "uchile", Name: "Universidad de Chile"}))
	f.utemAdmin = &models.Admin{ID: uuid.New(), Email: "admin@utem.test", Role: models.RoleInstitutionAdmin, InstitutionID: &f.utem.ID}
	f.utemProgram = must(f.programs.Create(ctx, f.super, f.utem.ID, ProgramInput{Name: "Ingeniería en Informática", Level: models.LevelUndergraduate}))
	f.masterProgram = must(f.programs.Create(ctx, f.super, f.utem.ID, ProgramInput{Name: "Magíster en Ciencia de Datos", Level: models.LevelMaster}))
	f.uchileProgram = must(f.programs.Create(ctx, f.super, f.uchile.ID, ProgramInput{Name: "Medicina", Level: models.LevelUndergraduate}))
	return f
}

func studentInput(program *models.Program, email string) StudentInput {
	return StudentInput{ProgramID: program.ID, Email: email, FullName: "Ana Pérez", EntryYear: 2024}
}

func TestInstitutionsOnlySuperAdminChangesThem(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if _, err := f.institutions.Create(ctx, f.utemAdmin, InstitutionInput{Code: "usach", Name: "Universidad de Santiago"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("create as institution admin: err = %v, want ErrForbidden", err)
	}
	if _, err := f.institutions.Update(ctx, f.utemAdmin, f.utem.ID, InstitutionInput{Code: "utem", Name: "Otro nombre"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("update as institution admin: err = %v, want ErrForbidden", err)
	}
	if err := f.institutions.Delete(ctx, f.utemAdmin, f.utem.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("delete as institution admin: err = %v, want ErrForbidden", err)
	}
}

func TestInstitutionsScope(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	list := must(f.institutions.List(ctx, f.utemAdmin))
	if len(list) != 1 || list[0].ID != f.utem.ID {
		t.Errorf("institution admin sees %v, want only utem", list)
	}
	if all := must(f.institutions.List(ctx, f.super)); len(all) != 2 {
		t.Errorf("super admin sees %d institutions, want 2", len(all))
	}
	if _, err := f.institutions.Get(ctx, f.utemAdmin, f.uchile.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("get another institution: err = %v, want ErrNotFound", err)
	}
}

func TestInstitutionValidation(t *testing.T) {
	f := newFixture(t)
	tests := map[string]InstitutionInput{
		"code with spaces":  {Code: "u tem", Name: "Nombre válido"},
		"code too short":    {Code: "a", Name: "Nombre válido"},
		"code upper accent": {Code: "Ñandú", Name: "Nombre válido"},
		"name too short":    {Code: "usach", Name: "U"},
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := f.institutions.Create(context.Background(), f.super, input); !errors.Is(err, ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
		})
	}

	if _, err := f.institutions.Create(context.Background(), f.super, InstitutionInput{Code: " UTEM ", Name: "Duplicada"}); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("duplicate code (case-insensitive): err = %v, want ErrConflict", err)
	}
}

func TestInstitutionCannotBeDeletedWhileInUse(t *testing.T) {
	f := newFixture(t)
	if err := f.institutions.Delete(context.Background(), f.super, f.utem.ID); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict (it has programs)", err)
	}
	empty := must(f.institutions.Create(context.Background(), f.super, InstitutionInput{Code: "vacia", Name: "Institución vacía"}))
	if err := f.institutions.Delete(context.Background(), f.super, empty.ID); err != nil {
		t.Errorf("delete empty institution: %v", err)
	}
}

func TestProgramsScope(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if _, err := f.programs.Create(ctx, f.utemAdmin, f.uchile.ID, ProgramInput{Name: "Derecho", Level: models.LevelUndergraduate}); !errors.Is(err, ErrForbidden) {
		t.Errorf("create in another institution: err = %v, want ErrForbidden", err)
	}
	if _, err := f.programs.Get(ctx, f.utemAdmin, f.uchileProgram.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("get a program of another institution: err = %v, want ErrNotFound", err)
	}
	if _, err := f.programs.List(ctx, f.utemAdmin, f.uchile.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("list programs of another institution: err = %v, want ErrNotFound", err)
	}
	if _, err := f.programs.Create(ctx, f.utemAdmin, f.utem.ID, ProgramInput{Name: "Programa", Level: "nope"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid level: err = %v, want ErrInvalid", err)
	}
	if _, err := f.programs.Create(ctx, f.super, uuid.New(), ProgramInput{Name: "Programa", Level: models.LevelMaster}); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("unknown institution: err = %v, want ErrNotFound", err)
	}
}

func TestLevelHelpers(t *testing.T) {
	if models.LevelUndergraduate.Postgraduate() {
		t.Error("undergraduate must not be postgraduate")
	}
	for _, level := range []models.Level{models.LevelMaster, models.LevelDoctorate, models.LevelOtherPostgraduate} {
		if !level.Postgraduate() {
			t.Errorf("%s must be postgraduate", level)
		}
	}
	if models.Level("x").Postgraduate() {
		t.Error("an invalid level must not be postgraduate")
	}
}

func TestStudentsCreate(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	student := must(f.students.Create(f.ctx(), f.utemAdmin, StudentInput{
		ProgramID: f.utemProgram.ID, Email: "  Ana.Perez@UTEM.test ", FullName: " Ana Pérez ", EntryYear: 2024,
	}))
	if student.Email != "ana.perez@utem.test" || student.FullName != "Ana Pérez" {
		t.Errorf("email/name not normalized: %q, %q", student.Email, student.FullName)
	}
	if student.InstitutionID != f.utem.ID || student.Status != models.StatusActive {
		t.Errorf("institution = %v, status = %q; want utem and active", student.InstitutionID, student.Status)
	}

	tests := []struct {
		name  string
		actor *models.Admin
		input StudentInput
		want  error
	}{
		{"duplicate email, any case", f.super, studentInput(f.uchileProgram, "ANA.PEREZ@utem.test"), repository.ErrConflict},
		{"program of another institution", f.utemAdmin, studentInput(f.uchileProgram, "x@uchile.test"), ErrForbidden},
		{"unknown program", f.super, StudentInput{ProgramID: uuid.New(), Email: "x@utem.test", FullName: "Ana", EntryYear: 2024}, ErrInvalid},
		{"invalid email", f.super, studentInput(f.utemProgram, "no-es-un-correo"), ErrInvalid},
		{"email with display name", f.super, studentInput(f.utemProgram, "Ana <ana@utem.test>"), ErrInvalid},
		{"email without domain dot", f.super, studentInput(f.utemProgram, "ana@localhost"), ErrInvalid},
		{"empty name", f.super, StudentInput{ProgramID: f.utemProgram.ID, Email: "n@utem.test", FullName: " ", EntryYear: 2024}, ErrInvalid},
		{"entry year in the past", f.super, StudentInput{ProgramID: f.utemProgram.ID, Email: "n@utem.test", FullName: "Nombre", EntryYear: 1800}, ErrInvalid},
		{"entry year too far ahead", f.super, StudentInput{ProgramID: f.utemProgram.ID, Email: "n@utem.test", FullName: "Nombre", EntryYear: 3000}, ErrInvalid},
		{"invalid status", f.super, StudentInput{ProgramID: f.utemProgram.ID, Email: "n@utem.test", FullName: "Nombre", EntryYear: 2024, Status: "otro"}, ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := f.students.Create(ctx, tt.actor, tt.input); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func (f *fixture) ctx() context.Context { return context.Background() }

func TestStudentsUpdateAndScope(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	student := must(f.students.Create(ctx, f.utemAdmin, studentInput(f.utemProgram, "ana@utem.test")))
	other := must(f.students.Create(ctx, f.super, studentInput(f.uchileProgram, "luis@uchile.test")))

	moved := must(f.students.Update(ctx, f.utemAdmin, student.ID, StudentInput{
		ProgramID: f.masterProgram.ID, Email: "ana@utem.test", FullName: "Ana Pérez Soto", EntryYear: 2025, Status: models.StatusOnLeave,
	}))
	if moved.ProgramID != f.masterProgram.ID || moved.Status != models.StatusOnLeave || moved.FullName != "Ana Pérez Soto" || !moved.CreatedAt.Equal(student.CreatedAt) {
		t.Errorf("update result unexpected: %+v", moved)
	}

	if _, err := f.students.Update(ctx, f.utemAdmin, student.ID, studentInput(f.uchileProgram, "ana@utem.test")); !errors.Is(err, ErrInvalid) {
		t.Errorf("move to another institution: err = %v, want ErrInvalid", err)
	}
	if _, err := f.students.Get(ctx, f.utemAdmin, other.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("get a student of another institution: err = %v, want ErrNotFound", err)
	}
	if _, err := f.students.Update(ctx, f.utemAdmin, other.ID, studentInput(f.utemProgram, "luis@uchile.test")); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("update a student of another institution: err = %v, want ErrNotFound", err)
	}
	if err := f.students.Delete(ctx, f.utemAdmin, other.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("delete a student of another institution: err = %v, want ErrNotFound", err)
	}
	if _, err := f.students.Update(ctx, f.super, student.ID, studentInput(f.utemProgram, "luis@uchile.test")); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("update to a taken email: err = %v, want ErrConflict", err)
	}
	if err := f.students.Delete(ctx, f.utemAdmin, student.ID); err != nil {
		t.Errorf("delete own student: %v", err)
	}
	if err := f.programs.Delete(ctx, f.super, f.uchileProgram.ID); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("delete a program with students: err = %v, want ErrConflict", err)
	}
}

func TestStudentsListScopeAndFilters(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	must(f.students.Create(ctx, f.super, StudentInput{ProgramID: f.utemProgram.ID, Email: "ana@utem.test", FullName: "Ana Pérez", EntryYear: 2022}))
	must(f.students.Create(ctx, f.super, StudentInput{ProgramID: f.masterProgram.ID, Email: "bruno@utem.test", FullName: "Bruno Díaz", EntryYear: 2023, Status: models.StatusGraduated}))
	must(f.students.Create(ctx, f.super, StudentInput{ProgramID: f.uchileProgram.ID, Email: "carla@uchile.test", FullName: "Carla Ruiz", EntryYear: 2021}))

	emails := func(actor *models.Admin, filter repository.StudentFilter) []string {
		var out []string
		for _, s := range must(f.students.List(ctx, actor, filter)) {
			out = append(out, s.Email)
		}
		return out
	}
	equal := func(got []string, want ...string) bool {
		if len(got) != len(want) {
			return false
		}
		for i := range got {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}

	if got := emails(f.super, repository.StudentFilter{}); len(got) != 3 {
		t.Errorf("super admin sees %v, want 3 students", got)
	}
	if got := emails(f.utemAdmin, repository.StudentFilter{}); !equal(got, "ana@utem.test", "bruno@utem.test") {
		t.Errorf("institution admin sees %v, want only utem students", got)
	}
	// Pedir otra institución no amplía el alcance.
	if got := emails(f.utemAdmin, repository.StudentFilter{InstitutionID: &f.uchile.ID}); !equal(got, "ana@utem.test", "bruno@utem.test") {
		t.Errorf("institution admin asking for another institution sees %v", got)
	}
	if got := emails(f.super, repository.StudentFilter{Level: models.LevelMaster}); !equal(got, "bruno@utem.test") {
		t.Errorf("level filter = %v", got)
	}
	if got := emails(f.super, repository.StudentFilter{Status: models.StatusGraduated}); !equal(got, "bruno@utem.test") {
		t.Errorf("status filter = %v", got)
	}
	if got := emails(f.super, repository.StudentFilter{Search: "RUIZ"}); !equal(got, "carla@uchile.test") {
		t.Errorf("search by name = %v", got)
	}
	if got := emails(f.super, repository.StudentFilter{Search: "ana@"}); !equal(got, "ana@utem.test") {
		t.Errorf("search by email prefix = %v", got)
	}
	if got := emails(f.super, repository.StudentFilter{Limit: 1, Offset: 1}); !equal(got, "bruno@utem.test") {
		t.Errorf("pagination = %v", got)
	}
	if _, err := f.students.List(ctx, f.super, repository.StudentFilter{Level: "x"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid level filter: err = %v, want ErrInvalid", err)
	}
	if _, err := f.students.List(ctx, f.super, repository.StudentFilter{Status: "x"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid status filter: err = %v, want ErrInvalid", err)
	}
}

func TestLookup(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	must(f.students.Create(ctx, f.super, studentInput(f.masterProgram, "ana@utem.test")))

	lookup := must(f.students.Lookup(ctx, " ANA@utem.test "))
	if lookup.Level != models.LevelMaster || lookup.InstitutionName != f.utem.Name || lookup.ProgramName != f.masterProgram.Name || lookup.Status != models.StatusActive {
		t.Errorf("lookup = %+v", lookup)
	}
	if _, err := f.students.Lookup(ctx, "nadie@utem.test"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("unknown email: err = %v, want ErrNotFound", err)
	}
	if _, err := f.students.Lookup(ctx, "basura"); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid email: err = %v, want ErrInvalid", err)
	}
}

func TestAuthenticate(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	must(f.admins.Create(ctx, f.super, AdminInput{Email: "Boss@UTEM.test", Role: models.RoleInstitutionAdmin, InstitutionID: &f.utem.ID}))

	tests := []struct {
		name     string
		identity *models.Identity
		wantErr  error
	}{
		{"registered and verified, any case", &models.Identity{UID: "u", Email: "boss@utem.TEST", EmailVerified: true}, nil},
		{"registered but email not verified", &models.Identity{UID: "u", Email: "boss@utem.test", EmailVerified: false}, ErrForbidden},
		{"verified but not an administrator", &models.Identity{UID: "u", Email: "otra@utem.test", EmailVerified: true}, ErrForbidden},
		{"no email in the token", &models.Identity{UID: "u", EmailVerified: true}, ErrForbidden},
		{"no identity", nil, ErrForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			admin, err := f.admins.Authenticate(ctx, tt.identity)
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && admin.Role != models.RoleInstitutionAdmin) {
				t.Errorf("admin = %v, err = %v, want %v", admin, err, tt.wantErr)
			}
		})
	}
}

func TestAdminsRules(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	tests := []struct {
		name  string
		actor *models.Admin
		input AdminInput
		want  error
	}{
		{"institution admin can't create admins", f.utemAdmin, AdminInput{Email: "a@utem.test", Role: models.RoleInstitutionAdmin, InstitutionID: &f.utem.ID}, ErrForbidden},
		{"institution admin needs an institution", f.super, AdminInput{Email: "a@utem.test", Role: models.RoleInstitutionAdmin}, ErrInvalid},
		{"super admin has no institution", f.super, AdminInput{Email: "a@utem.test", Role: models.RoleSuperAdmin, InstitutionID: &f.utem.ID}, ErrInvalid},
		{"unknown institution", f.super, AdminInput{Email: "a@utem.test", Role: models.RoleInstitutionAdmin, InstitutionID: ptr(uuid.New())}, ErrInvalid},
		{"unknown role", f.super, AdminInput{Email: "a@utem.test", Role: "dios"}, ErrInvalid},
		{"invalid email", f.super, AdminInput{Email: "x", Role: models.RoleSuperAdmin}, ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := f.admins.Create(ctx, tt.actor, tt.input); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}

	created := must(f.admins.Create(ctx, f.super, AdminInput{Email: "a@utem.test", Role: models.RoleInstitutionAdmin, InstitutionID: &f.utem.ID}))
	if _, err := f.admins.Create(ctx, f.super, AdminInput{Email: "A@utem.test", Role: models.RoleSuperAdmin}); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("duplicate admin: err = %v, want ErrConflict", err)
	}
	if _, err := f.admins.List(ctx, f.utemAdmin); !errors.Is(err, ErrForbidden) {
		t.Errorf("list as institution admin: err = %v, want ErrForbidden", err)
	}
	if err := f.admins.Delete(ctx, f.super, f.super.ID); !errors.Is(err, ErrInvalid) {
		t.Errorf("delete yourself: err = %v, want ErrInvalid", err)
	}
	if err := f.admins.Delete(ctx, f.utemAdmin, created.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("delete as institution admin: err = %v, want ErrForbidden", err)
	}
	if err := f.admins.Delete(ctx, f.super, created.ID); err != nil {
		t.Errorf("delete admin: %v", err)
	}
}

func ptr[T any](v T) *T { return &v }

func TestBootstrapIsIdempotent(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if err := f.admins.Bootstrap(ctx, "Primero@Laurelin.test"); err != nil {
			t.Fatalf("bootstrap #%d: %v", i+1, err)
		}
	}
	admin := must(f.mem.Admins().GetByEmail(ctx, "primero@laurelin.test"))
	if admin.Role != models.RoleSuperAdmin || admin.InstitutionID != nil {
		t.Errorf("bootstrap admin = %+v, want super_admin without institution", admin)
	}
	if err := f.admins.Bootstrap(ctx, "no-es-correo"); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid bootstrap email: err = %v, want ErrInvalid", err)
	}
}

func TestMutationsAreAuditedWithoutPersonalData(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	before := len(f.mem.Audit)

	student := must(f.students.Create(ctx, f.utemAdmin, studentInput(f.utemProgram, "ana@utem.test")))
	must(f.students.Update(ctx, f.utemAdmin, student.ID, studentInput(f.utemProgram, "ana@utem.test")))
	if err := f.students.Delete(ctx, f.utemAdmin, student.ID); err != nil {
		t.Fatal(err)
	}
	// Una lectura y un intento fallido no dejan rastro.
	_, _ = f.students.Get(ctx, f.utemAdmin, student.ID)
	_, _ = f.students.Create(ctx, f.utemAdmin, studentInput(f.uchileProgram, "x@uchile.test"))

	entries := f.mem.Audit[before:]
	var actions []string
	for _, e := range entries {
		actions = append(actions, e.Action)
		if e.ActorAdminID != f.utemAdmin.ID || e.EntityID != student.ID || e.InstitutionID == nil || *e.InstitutionID != f.utem.ID {
			t.Errorf("audit entry = %+v", e)
		}
	}
	want := []string{"student.create", "student.update", "student.delete"}
	if len(actions) != len(want) {
		t.Fatalf("actions = %v, want %v", actions, want)
	}
	for i := range want {
		if actions[i] != want[i] {
			t.Errorf("actions = %v, want %v", actions, want)
		}
	}
}

func TestClampPage(t *testing.T) {
	tests := []struct{ limit, offset, wantLimit, wantOffset int }{
		{0, 0, DefaultPageSize, 0},
		{-5, -3, DefaultPageSize, 0},
		{10, 7, 10, 7},
		{100000, 0, MaxPageSize, 0},
	}
	for _, tt := range tests {
		if l, o := clampPage(tt.limit, tt.offset); l != tt.wantLimit || o != tt.wantOffset {
			t.Errorf("clampPage(%d, %d) = %d, %d; want %d, %d", tt.limit, tt.offset, l, o, tt.wantLimit, tt.wantOffset)
		}
	}
}
