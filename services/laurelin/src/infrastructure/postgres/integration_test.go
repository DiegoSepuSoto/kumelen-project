//go:build integration

// Prueba los repositorios contra un PostgreSQL real. Borra las tablas de Laurelin de esa base:
// usa solo una base desechable. Variable: LAURELIN_TEST_DATABASE_URL, por ejemplo
//
//	podman run -d --rm -e POSTGRES_PASSWORD=test -e POSTGRES_DB=laurelin -p 55433:5432 postgres:17
//	LAURELIN_TEST_DATABASE_URL=postgres://postgres:test@localhost:55433/laurelin go test -tags integration ./src/infrastructure/postgres/
package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/models"
	"laurelin/src/domain/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("LAURELIN_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("LAURELIN_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS audit_log, students, admins, programs, institutions, schema_migrations CASCADE"); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestMigrateIsIdempotentAndSafeConcurrently(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- Migrate(ctx, pool)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent migrate: %v", err)
		}
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate again: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&count); err != nil || count != 1 {
		t.Errorf("applied migrations = %d (err %v), want 1", count, err)
	}
}

// TestDomainRulesOnRealDatabase ejecuta los casos de uso reales sobre PostgreSQL.
func TestDomainRulesOnRealDatabase(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}

	institutionsRepo, programsRepo := NewInstitutionRepository(pool), NewProgramRepository(pool)
	studentsRepo, adminsRepo, auditRepo := NewStudentRepository(pool), NewAdminRepository(pool), NewAuditRepository(pool)
	institutions := usecase.NewInstitutions(institutionsRepo, auditRepo)
	programs := usecase.NewPrograms(programsRepo, institutionsRepo, auditRepo)
	students := usecase.NewStudents(studentsRepo, programsRepo, auditRepo)
	admins := usecase.NewAdmins(adminsRepo, institutionsRepo, auditRepo)

	if err := admins.Bootstrap(ctx, "Root@Laurelin.test"); err != nil {
		t.Fatal(err)
	}
	if err := admins.Bootstrap(ctx, "root@laurelin.test"); err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}
	root, err := admins.Authenticate(ctx, &models.Identity{UID: "1", Email: "ROOT@laurelin.test", EmailVerified: true})
	if err != nil || !root.IsSuper() {
		t.Fatalf("authenticate root = %+v, %v", root, err)
	}

	utem, err := institutions.Create(ctx, root, usecase.InstitutionInput{Code: "utem", Name: "Universidad Tecnológica Metropolitana"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := institutions.Create(ctx, root, usecase.InstitutionInput{Code: "utem", Name: "Duplicada"}); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("duplicate code: err = %v, want ErrConflict", err)
	}
	uchile, _ := institutions.Create(ctx, root, usecase.InstitutionInput{Code: "uchile", Name: "Universidad de Chile"})

	bachelor, err := programs.Create(ctx, root, utem.ID, usecase.ProgramInput{Name: "Ingeniería Civil en Informática", Level: models.LevelUndergraduate})
	if err != nil {
		t.Fatal(err)
	}
	master, _ := programs.Create(ctx, root, utem.ID, usecase.ProgramInput{Name: "Magíster en Datos", Level: models.LevelMaster})
	medicine, _ := programs.Create(ctx, root, uchile.ID, usecase.ProgramInput{Name: "Medicina", Level: models.LevelUndergraduate})
	if _, err := programs.Create(ctx, root, utem.ID, usecase.ProgramInput{Name: "Magíster en Datos", Level: models.LevelMaster}); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("duplicate program: err = %v, want ErrConflict", err)
	}

	utemAdmin, err := admins.Create(ctx, root, usecase.AdminInput{Email: "admin@utem.test", Role: models.RoleInstitutionAdmin, InstitutionID: &utem.ID})
	if err != nil {
		t.Fatal(err)
	}

	ana, err := students.Create(ctx, utemAdmin, usecase.StudentInput{ProgramID: bachelor.ID, Email: "Ana@utem.test", FullName: "Ana Pérez", EntryYear: 2022})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := students.Create(ctx, root, usecase.StudentInput{ProgramID: medicine.ID, Email: "ANA@utem.test", FullName: "Otra Ana", EntryYear: 2022}); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("duplicate email: err = %v, want ErrConflict", err)
	}
	students.Create(ctx, root, usecase.StudentInput{ProgramID: master.ID, Email: "bruno@utem.test", FullName: "Bruno 100%_Díaz", EntryYear: 2023, Status: models.StatusGraduated})
	students.Create(ctx, root, usecase.StudentInput{ProgramID: medicine.ID, Email: "carla@uchile.test", FullName: "Carla Ruiz", EntryYear: 2021})

	names := func(actor *models.Admin, filter repository.StudentFilter) string {
		list, err := students.List(ctx, actor, filter)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, s := range list {
			out = append(out, strings.Split(s.Email, "@")[0])
		}
		return strings.Join(out, ",")
	}
	checks := []struct {
		name   string
		actor  *models.Admin
		filter repository.StudentFilter
		want   string
	}{
		{"super admin sees all, ordered by name", root, repository.StudentFilter{}, "ana,bruno,carla"},
		{"institution admin only sees its own", utemAdmin, repository.StudentFilter{}, "ana,bruno"},
		{"institution admin can't widen the scope", utemAdmin, repository.StudentFilter{InstitutionID: &uchile.ID}, "ana,bruno"},
		{"level", root, repository.StudentFilter{Level: models.LevelMaster}, "bruno"},
		{"status", root, repository.StudentFilter{Status: models.StatusGraduated}, "bruno"},
		{"program", root, repository.StudentFilter{ProgramID: &medicine.ID}, "carla"},
		{"search by name, any case", root, repository.StudentFilter{Search: "RUIZ"}, "carla"},
		{"search by email prefix", root, repository.StudentFilter{Search: "ana@"}, "ana"},
		{"percent is literal, not a wildcard", root, repository.StudentFilter{Search: "100%"}, "bruno"},
		{"a lone percent doesn't match everything", root, repository.StudentFilter{Search: "%"}, "bruno"},
		{"underscore is literal", root, repository.StudentFilter{Search: "a_a"}, ""},
		{"pagination", root, repository.StudentFilter{Limit: 1, Offset: 1}, "bruno"},
		{"combined", root, repository.StudentFilter{Level: models.LevelUndergraduate, Status: models.StatusActive}, "ana,carla"},
	}
	for _, c := range checks {
		if got := names(c.actor, c.filter); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}

	lookup, err := students.Lookup(ctx, "Bruno@UTEM.test")
	if err != nil || lookup.Level != models.LevelMaster || lookup.ProgramName != "Magíster en Datos" || lookup.InstitutionName != utem.Name || lookup.Status != models.StatusGraduated {
		t.Errorf("lookup = %+v, %v", lookup, err)
	}
	if _, err := students.Lookup(ctx, "nadie@utem.test"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("unknown lookup: err = %v", err)
	}

	moved, err := students.Update(ctx, utemAdmin, ana.ID, usecase.StudentInput{ProgramID: master.ID, Email: "ana@utem.test", FullName: "Ana Pérez Soto", EntryYear: 2022, Status: models.StatusOnLeave})
	if err != nil || moved.ProgramID != master.ID || moved.Status != models.StatusOnLeave {
		t.Errorf("update = %+v, %v", moved, err)
	}
	if _, err := students.Update(ctx, root, ana.ID, usecase.StudentInput{ProgramID: master.ID, Email: "bruno@utem.test", FullName: "Ana", EntryYear: 2022}); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("update to a taken email: err = %v, want ErrConflict", err)
	}

	// Nada que se use se puede borrar (claves foráneas), y lo vacío sí.
	if err := programs.Delete(ctx, root, master.ID); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("delete program with students: err = %v, want ErrConflict", err)
	}
	if err := institutions.Delete(ctx, root, utem.ID); !errors.Is(err, repository.ErrConflict) {
		t.Errorf("delete institution in use: err = %v, want ErrConflict", err)
	}
	if err := students.Delete(ctx, utemAdmin, ana.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := students.Get(ctx, root, ana.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("get deleted student: err = %v, want ErrNotFound", err)
	}
	if err := admins.Delete(ctx, root, uuid.New()); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("delete unknown admin: err = %v, want ErrNotFound", err)
	}

	// Las reglas de la base protegen aunque se salte la aplicación.
	if _, err := pool.Exec(ctx, "INSERT INTO students (id, institution_id, program_id, email, full_name, entry_year, status, created_at, updated_at) VALUES ($1,$2,$3,'Mayus@X.cl','n',2020,'active',now(),now())", uuid.New(), utem.ID, bachelor.ID); err == nil {
		t.Error("the database accepted an email with uppercase letters")
	}
	if _, err := pool.Exec(ctx, "INSERT INTO admins (id, email, role, institution_id, created_at) VALUES ($1,'x@y.cl','institution_admin',NULL,now())", uuid.New()); err == nil {
		t.Error("the database accepted an institution_admin without institution")
	}

	// La auditoría guarda ids y acciones, nunca correos ni nombres.
	rows, err := pool.Query(ctx, "SELECT action FROM audit_log ORDER BY at")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	actions := 0
	for rows.Next() {
		actions++
	}
	if actions < 10 {
		t.Errorf("audit rows = %d, want at least 10", actions)
	}
	var leaks int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM audit_log WHERE action ~ '@'").Scan(&leaks); err != nil || leaks != 0 {
		t.Errorf("audit leaks = %d (err %v)", leaks, err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Error(err)
	}
}
