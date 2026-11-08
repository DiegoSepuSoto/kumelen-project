package middleware

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"iluvatar/src/domain/models"
	"iluvatar/src/domain/repository"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeVerifier struct {
	identity *models.Identity
	err      error
	received []string
}

func (f *fakeVerifier) Verify(_ context.Context, idToken string) (*models.Identity, error) {
	f.received = append(f.received, idToken)
	return f.identity, f.err
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		verifier     *fakeVerifier
		wantStatus   int
		wantVerified []string
		wantUID      string
	}{
		{"valid token", "Bearer good", &fakeVerifier{identity: &models.Identity{UID: "uid-1"}}, http.StatusOK, []string{"good"}, "uid-1"},
		{"surrounding spaces are trimmed", "Bearer  good ", &fakeVerifier{identity: &models.Identity{UID: "uid-1"}}, http.StatusOK, []string{"good"}, "uid-1"},
		{"missing header", "", &fakeVerifier{}, http.StatusUnauthorized, nil, ""},
		{"token without the Bearer prefix", "good", &fakeVerifier{}, http.StatusUnauthorized, nil, ""},
		{"lowercase scheme", "bearer good", &fakeVerifier{}, http.StatusUnauthorized, nil, ""},
		{"Bearer with no token", "Bearer ", &fakeVerifier{}, http.StatusUnauthorized, nil, ""},
		{"rejected token", "Bearer bad", &fakeVerifier{err: repository.ErrInvalidToken}, http.StatusUnauthorized, []string{"bad"}, ""},
		{"verifier failure", "Bearer any", &fakeVerifier{err: errors.New("keys unavailable")}, http.StatusUnauthorized, []string{"any"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUID string
			nextCalled := false
			handler := RequireAuth(tt.verifier)(func(c echo.Context) error {
				nextCalled = true
				gotUID = UID(c)
				return c.NoContent(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				request.Header.Set(echo.HeaderAuthorization, tt.header)
			}
			recorder := httptest.NewRecorder()

			if err := handler(echo.New().NewContext(request, recorder)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if recorder.Code != tt.wantStatus || nextCalled != (tt.wantStatus == http.StatusOK) {
				t.Errorf("status = %d, next = %v; want %d", recorder.Code, nextCalled, tt.wantStatus)
			}
			if gotUID != tt.wantUID {
				t.Errorf("uid in context = %q, want %q", gotUID, tt.wantUID)
			}
			if len(tt.verifier.received) != len(tt.wantVerified) || (len(tt.wantVerified) > 0 && tt.verifier.received[0] != tt.wantVerified[0]) {
				t.Errorf("verified tokens = %v, want %v", tt.verifier.received, tt.wantVerified)
			}
		})
	}
}

func TestUIDWithoutAuthentication(t *testing.T) {
	if uid := UID(echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())); uid != "" {
		t.Errorf("UID() = %q, want empty", uid)
	}
}
