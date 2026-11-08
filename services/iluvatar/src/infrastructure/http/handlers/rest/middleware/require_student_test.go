package middleware

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"iluvatar/src/domain/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeDirectory struct {
	standing *models.StudentStanding
	err      error
	asked    []string
}

func (f *fakeDirectory) Lookup(_ context.Context, email string) (*models.StudentStanding, error) {
	f.asked = append(f.asked, email)
	return f.standing, f.err
}

func TestRequireStudent(t *testing.T) {
	verified := &models.Identity{UID: "u", Email: "ana@utem.cl", EmailVerified: true}
	tests := []struct {
		name       string
		identity   *models.Identity
		directory  *fakeDirectory
		wantStatus int
		wantAsked  int
	}{
		{"active student", verified, &fakeDirectory{standing: &models.StudentStanding{IsStudent: true}}, http.StatusOK, 1},
		{"not a student", verified, &fakeDirectory{standing: &models.StudentStanding{}}, http.StatusForbidden, 1},
		{"unverified email is never looked up", &models.Identity{UID: "u", Email: "ana@utem.cl"}, &fakeDirectory{}, http.StatusForbidden, 0},
		{"account without email", &models.Identity{UID: "u", EmailVerified: true}, &fakeDirectory{}, http.StatusForbidden, 0},
		{"laurelin down fails closed", verified, &fakeDirectory{err: errors.New("connection refused")}, http.StatusServiceUnavailable, 1},
		{"no identity (RequireAuth missing)", nil, &fakeDirectory{}, http.StatusUnauthorized, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			handler := RequireStudent(tt.directory)(func(c echo.Context) error {
				nextCalled = true
				return c.NoContent(http.StatusOK)
			})
			c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
			if tt.identity != nil {
				c.Set(identityKey, tt.identity)
			}

			if err := handler(c); err != nil {
				t.Fatal(err)
			}
			recorder := c.Response().Writer.(*httptest.ResponseRecorder)
			if recorder.Code != tt.wantStatus || nextCalled != (tt.wantStatus == http.StatusOK) || len(tt.directory.asked) != tt.wantAsked {
				t.Errorf("status = %d, next = %v, lookups = %d", recorder.Code, nextCalled, len(tt.directory.asked))
			}
			if strings.Contains(recorder.Body.String(), "utem.cl") {
				t.Errorf("the response leaks the email: %s", recorder.Body)
			}
		})
	}
}

func TestRequireAuthLeavesTheIdentityInTheContext(t *testing.T) {
	verifier := &fakeVerifier{identity: &models.Identity{UID: "u", Email: "ana@utem.cl", EmailVerified: true}}
	var got *models.Identity
	handler := RequireAuth(verifier)(func(c echo.Context) error {
		got = Identity(c)
		return c.NoContent(http.StatusOK)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(echo.HeaderAuthorization, "Bearer good")

	if err := handler(echo.New().NewContext(request, httptest.NewRecorder())); err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Email != "ana@utem.cl" || !got.EmailVerified {
		t.Errorf("identity = %+v", got)
	}
	if Identity(echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())) != nil {
		t.Error("Identity() without authentication must be nil")
	}
}
