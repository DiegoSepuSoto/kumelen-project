package health

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func get(e *echo.Echo, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestHealthDoesNotDependOnOtherServices(t *testing.T) {
	t.Setenv("APP_VERSION", "1.2.3")
	t.Setenv("ENV", "production")
	e := echo.New()
	NewHealthHandler(e, Check{Name: "cms", Run: func(context.Context) error { return errors.New("down") }})

	recorder := get(e, "/health")

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"version":"1.2.3"`) {
		t.Errorf("status = %d, body = %s", recorder.Code, recorder.Body)
	}
}

func TestReady(t *testing.T) {
	ok := func(context.Context) error { return nil }
	fail := func(context.Context) error { return errors.New("secret internal address 10.0.0.7 refused") }

	tests := []struct {
		name       string
		checks     []Check
		wantStatus int
		wantChecks map[string]string
	}{
		{"no dependencies", nil, http.StatusOK, map[string]string{}},
		{"all up", []Check{{"cms", ok}, {"other", ok}}, http.StatusOK, map[string]string{"cms": "UP", "other": "UP"}},
		{"one down", []Check{{"cms", fail}, {"other", ok}}, http.StatusServiceUnavailable, map[string]string{"cms": "DOWN", "other": "UP"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			NewHealthHandler(e, tt.checks...)

			recorder := get(e, "/ready")

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			var body readiness
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || !reflect.DeepEqual(body.Checks, tt.wantChecks) {
				t.Errorf("checks = %v (%v), want %v", body.Checks, err, tt.wantChecks)
			}
			if strings.Contains(recorder.Body.String(), "10.0.0.7") {
				t.Error("the failure reason must not reach the client")
			}
		})
	}
}

func TestReadyGivesUpOnASlowDependency(t *testing.T) {
	e := echo.New()
	NewHealthHandler(e, Check{Name: "cms", Run: func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	}})

	started := time.Now()
	recorder := get(e, "/ready")

	if recorder.Code != http.StatusServiceUnavailable || time.Since(started) > 4*time.Second {
		t.Errorf("status = %d after %v, want 503 within the check timeout", recorder.Code, time.Since(started))
	}
}
