package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"laurelin/src/application/usecase"
	"laurelin/src/domain/models"
	customMiddleware "laurelin/src/infrastructure/http/handlers/rest/middleware"
	serviceHandler "laurelin/src/infrastructure/http/handlers/rest/service"
	v1 "laurelin/src/infrastructure/http/handlers/rest/v1"
	"laurelin/src/shared/testdoubles"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const serviceToken = "service-secret"

type env struct {
	e   *echo.Echo
	mem *testdoubles.Memory
}

// newEnv monta las rutas reales con los casos de uso reales sobre repositorios en memoria.
// Tokens de prueba: "root" (super_admin), "utem-admin" (institution_admin de utem), "stranger" (cuenta sin permisos)
// y "unverified" (correo de administrador sin verificar).
func newEnv(t *testing.T) *env {
	t.Helper()
	mem := testdoubles.NewMemory()
	institutions := usecase.NewInstitutions(mem.Institutions(), mem.AuditLog())
	programs := usecase.NewPrograms(mem.Programs(), mem.Institutions(), mem.AuditLog())
	students := usecase.NewStudents(mem.Students(), mem.Programs(), mem.AuditLog())
	admins := usecase.NewAdmins(mem.Admins(), mem.Institutions(), mem.AuditLog())

	if err := admins.Bootstrap(context.Background(), "root@laurelin.test"); err != nil {
		t.Fatal(err)
	}
	root, _ := mem.Admins().GetByEmail(context.Background(), "root@laurelin.test")
	utem, err := institutions.Create(context.Background(), root, usecase.InstitutionInput{Code: "utem", Name: "Universidad Tecnológica Metropolitana"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admins.Create(context.Background(), root, usecase.AdminInput{Email: "admin@utem.test", Role: models.RoleInstitutionAdmin, InstitutionID: &utem.ID}); err != nil {
		t.Fatal(err)
	}

	verifier := testdoubles.Verifier{
		"root":       {UID: "1", Email: "root@laurelin.test", EmailVerified: true},
		"utem-admin": {UID: "2", Email: "admin@utem.test", EmailVerified: true},
		"stranger":   {UID: "3", Email: "otra@persona.test", EmailVerified: true},
		"unverified": {UID: "4", Email: "root@laurelin.test", EmailVerified: false},
	}
	e := echo.New()
	v1.Register(e, customMiddleware.RequireAdmin(verifier, admins), v1.Dependencies{
		Institutions: institutions, Programs: programs, Students: students, Admins: admins,
	})
	serviceHandler.Register(e, customMiddleware.RequireServiceToken(serviceToken), students)
	return &env{e: e, mem: mem}
}

func (n *env) do(method, path, token, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != "" {
		request.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	n.e.ServeHTTP(recorder, request)
	return recorder
}

func decode[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
		t.Fatalf("response is not the expected JSON: %v\n%s", err, recorder.Body)
	}
	return value
}

func (n *env) expect(t *testing.T, want int, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, want, recorder.Body)
	}
}

func TestAuthentication(t *testing.T) {
	n := newEnv(t)
	tests := []struct {
		name  string
		token string
		want  int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"invalid token", "garbage", http.StatusUnauthorized},
		{"valid account that administers nothing", "stranger", http.StatusForbidden},
		{"administrator with unverified email", "unverified", http.StatusForbidden},
		{"administrator", "utem-admin", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n.expect(t, tt.want, n.do(http.MethodGet, "/v1/me", tt.token, ""))
		})
	}

	for _, route := range []string{"/v1/institutions", "/v1/students", "/v1/admins", "/v1/me"} {
		if code := n.do(http.MethodGet, route, "", "").Code; code != http.StatusUnauthorized {
			t.Errorf("GET %s without token = %d, want 401", route, code)
		}
	}
}

func TestStudentCRUDFlow(t *testing.T) {
	n := newEnv(t)

	institutions := decode[struct{ Items []models.Institution }](t, n.do(http.MethodGet, "/v1/institutions", "root", ""))
	utem := institutions.Items[0]

	created := n.do(http.MethodPost, "/v1/institutions/"+utem.ID.String()+"/programs", "utem-admin", `{"name":"Magíster en Informática","level":"master"}`)
	n.expect(t, http.StatusCreated, created)
	program := decode[models.Program](t, created)

	body := `{"program_id":"` + program.ID.String() + `","email":"Ana@UTEM.test","full_name":"Ana Pérez","entry_year":2024}`
	createdStudent := n.do(http.MethodPost, "/v1/students", "utem-admin", body)
	n.expect(t, http.StatusCreated, createdStudent)
	student := decode[models.Student](t, createdStudent)
	if student.Email != "ana@utem.test" || student.Status != models.StatusActive || student.InstitutionID != utem.ID {
		t.Fatalf("student = %+v", student)
	}

	n.expect(t, http.StatusConflict, n.do(http.MethodPost, "/v1/students", "utem-admin", body))

	list := decode[struct {
		Items         []models.Student
		Limit, Offset int
	}](t, n.do(http.MethodGet, "/v1/students?level=master&q=ana", "utem-admin", ""))
	if len(list.Items) != 1 || list.Limit != usecase.DefaultPageSize || list.Offset != 0 {
		t.Errorf("list = %+v", list)
	}

	update := `{"program_id":"` + program.ID.String() + `","email":"ana@utem.test","full_name":"Ana Pérez Soto","entry_year":2024,"status":"graduated"}`
	updated := n.do(http.MethodPut, "/v1/students/"+student.ID.String(), "utem-admin", update)
	n.expect(t, http.StatusOK, updated)
	if got := decode[models.Student](t, updated); got.Status != models.StatusGraduated || got.FullName != "Ana Pérez Soto" {
		t.Errorf("updated = %+v", got)
	}

	n.expect(t, http.StatusOK, n.do(http.MethodGet, "/v1/students/"+student.ID.String(), "root", ""))
	n.expect(t, http.StatusNoContent, n.do(http.MethodDelete, "/v1/students/"+student.ID.String(), "utem-admin", ""))
	n.expect(t, http.StatusNotFound, n.do(http.MethodGet, "/v1/students/"+student.ID.String(), "utem-admin", ""))
}

func TestErrorMapping(t *testing.T) {
	n := newEnv(t)
	random := uuid.New().String()

	tests := []struct {
		name                      string
		method, path, token, body string
		want                      int
	}{
		{"malformed id", http.MethodGet, "/v1/students/not-a-uuid", "root", "", http.StatusBadRequest},
		{"malformed filter", http.MethodGet, "/v1/students?program_id=zzz", "root", "", http.StatusBadRequest},
		{"invalid filter value", http.MethodGet, "/v1/students?level=nope", "root", "", http.StatusBadRequest},
		{"malformed body", http.MethodPost, "/v1/institutions", "root", "{", http.StatusBadRequest},
		{"invalid body values", http.MethodPost, "/v1/institutions", "root", `{"code":"X y","name":"n"}`, http.StatusBadRequest},
		{"unknown student", http.MethodGet, "/v1/students/" + random, "root", "", http.StatusNotFound},
		{"unknown program in student", http.MethodPost, "/v1/students", "root", `{"program_id":"` + random + `","email":"a@b.cl","full_name":"Ana","entry_year":2024}`, http.StatusBadRequest},
		{"institution admin creates institution", http.MethodPost, "/v1/institutions", "utem-admin", `{"code":"usach","name":"Universidad de Santiago"}`, http.StatusForbidden},
		{"institution admin lists admins", http.MethodGet, "/v1/admins", "utem-admin", "", http.StatusForbidden},
		{"institution admin reads unknown institution", http.MethodGet, "/v1/institutions/" + random, "utem-admin", "", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n.expect(t, tt.want, n.do(tt.method, tt.path, tt.token, tt.body))
		})
	}
}

func TestInstitutionAndAdminManagement(t *testing.T) {
	n := newEnv(t)

	created := n.do(http.MethodPost, "/v1/institutions", "root", `{"code":"usach","name":"Universidad de Santiago"}`)
	n.expect(t, http.StatusCreated, created)
	usach := decode[models.Institution](t, created)
	n.expect(t, http.StatusConflict, n.do(http.MethodPost, "/v1/institutions", "root", `{"code":"usach","name":"Otra"}`))
	n.expect(t, http.StatusOK, n.do(http.MethodPut, "/v1/institutions/"+usach.ID.String(), "root", `{"code":"usach","name":"Universidad de Santiago de Chile"}`))

	admin := n.do(http.MethodPost, "/v1/admins", "root", `{"email":"jefe@usach.test","role":"institution_admin","institution_id":"`+usach.ID.String()+`"}`)
	n.expect(t, http.StatusCreated, admin)
	n.expect(t, http.StatusConflict, n.do(http.MethodDelete, "/v1/institutions/"+usach.ID.String(), "root", ""))

	adminID := decode[models.Admin](t, admin).ID.String()
	n.expect(t, http.StatusNoContent, n.do(http.MethodDelete, "/v1/admins/"+adminID, "root", ""))
	n.expect(t, http.StatusNoContent, n.do(http.MethodDelete, "/v1/institutions/"+usach.ID.String(), "root", ""))

	me := decode[models.Admin](t, n.do(http.MethodGet, "/v1/me", "root", ""))
	n.expect(t, http.StatusBadRequest, n.do(http.MethodDelete, "/v1/admins/"+me.ID.String(), "root", ""))
}

func TestLookup(t *testing.T) {
	n := newEnv(t)
	utem := decode[struct{ Items []models.Institution }](t, n.do(http.MethodGet, "/v1/institutions", "root", "")).Items[0]
	program := decode[models.Program](t, n.do(http.MethodPost, "/v1/institutions/"+utem.ID.String()+"/programs", "root", `{"name":"Ingeniería Civil","level":"undergraduate"}`))
	n.expect(t, http.StatusCreated, n.do(http.MethodPost, "/v1/students", "root",
		`{"program_id":"`+program.ID.String()+`","email":"ana@utem.test","full_name":"Ana Pérez","entry_year":2024}`))
	n.expect(t, http.StatusCreated, n.do(http.MethodPost, "/v1/students", "root",
		`{"program_id":"`+program.ID.String()+`","email":"egresado@utem.test","full_name":"Beto Gómez","entry_year":2015,"status":"graduated"}`))

	post := func(token, body string) *httptest.ResponseRecorder {
		return n.do(http.MethodPost, "/internal/v1/students/lookup", token, body)
	}

	found := post(serviceToken, `{"email":"ANA@utem.test"}`)
	n.expect(t, http.StatusOK, found)
	got := decode[map[string]any](t, found)
	if got["is_student"] != true || got["level"] != "undergraduate" || got["institution_name"] != utem.Name || got["program_name"] != "Ingeniería Civil" {
		t.Errorf("lookup = %v", got)
	}
	if _, leaks := got["email"]; leaks {
		t.Error("lookup must not echo the email")
	}
	if _, leaks := got["full_name"]; leaks {
		t.Error("lookup must not return the name")
	}

	graduated := decode[map[string]any](t, post(serviceToken, `{"email":"egresado@utem.test"}`))
	if graduated["is_student"] != false || graduated["status"] != "graduated" {
		t.Errorf("graduated lookup = %v", graduated)
	}

	n.expect(t, http.StatusNotFound, post(serviceToken, `{"email":"nadie@utem.test"}`))
	n.expect(t, http.StatusBadRequest, post(serviceToken, `{"email":"basura"}`))
	n.expect(t, http.StatusBadRequest, post(serviceToken, `{`))
	n.expect(t, http.StatusUnauthorized, post("", `{"email":"ana@utem.test"}`))
	n.expect(t, http.StatusUnauthorized, post("otro-token", `{"email":"ana@utem.test"}`))
	// Un ID token de administrador no sirve como token de servicio, ni al revés.
	n.expect(t, http.StatusUnauthorized, post("root", `{"email":"ana@utem.test"}`))
	n.expect(t, http.StatusUnauthorized, n.do(http.MethodGet, "/v1/me", serviceToken, ""))
}

func TestServiceTokenNotConfigured(t *testing.T) {
	e := echo.New()
	serviceHandler.Register(e, customMiddleware.RequireServiceToken(""), nil)
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/students/lookup", strings.NewReader(`{"email":"a@b.cl"}`))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer ")
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503 so the route is never open without a token", recorder.Code)
	}
}
