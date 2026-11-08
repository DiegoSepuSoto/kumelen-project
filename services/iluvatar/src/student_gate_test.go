package main

import (
	"github.com/labstack/echo/v4"
	"iluvatar/src/infrastructure/http/server"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStudentGateConfiguration(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenFile, []byte("  from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	production, local := server.Config{Environment: "production"}, server.Config{Environment: "local"}

	tests := []struct {
		name        string
		config      server.Config
		env         map[string]string
		wantGate    bool
		wantErrPart string
	}{
		{"local without Laurelin: disabled", local, nil, false, ""},
		{"production without Laurelin: refuses to start", production, nil, false, "LAURELIN_URL is required"},
		{"production with the explicit opt-out", production, map[string]string{"STUDENT_CHECK_DISABLED": "true"}, false, ""},
		{"url and token", local, map[string]string{"LAURELIN_URL": "http://laurelin:8090", "LAURELIN_TOKEN": "t"}, true, ""},
		{"url and token file", production, map[string]string{"LAURELIN_URL": "https://laurelin.example", "LAURELIN_TOKEN_FILE": tokenFile}, true, ""},
		{"url without token", local, map[string]string{"LAURELIN_URL": "http://laurelin:8090"}, false, "LAURELIN_TOKEN"},
		{"url without scheme", local, map[string]string{"LAURELIN_URL": "laurelin:8090", "LAURELIN_TOKEN": "t"}, false, "must start with"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, name := range []string{"LAURELIN_URL", "LAURELIN_TOKEN", "LAURELIN_TOKEN_FILE", "STUDENT_CHECK_DISABLED"} {
				t.Setenv(name, "")
			}
			for name, value := range tt.env {
				t.Setenv(name, value)
			}

			gate, err := newStudentGate(tt.config)

			if tt.wantErrPart != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Fatalf("err = %v, want it to contain %q", err, tt.wantErrPart)
				}
				return
			}
			if err != nil || (gate != nil) != tt.wantGate {
				t.Errorf("gate = %v, err = %v; want gate %v", gate != nil, err, tt.wantGate)
			}
		})
	}
}

func TestSecretFromEnv(t *testing.T) {
	file := filepath.Join(t.TempDir(), "secret")
	os.WriteFile(file, []byte("from-file\n"), 0o600)

	t.Setenv("X_SECRET", "direct")
	t.Setenv("X_SECRET_FILE", file)
	if got := secretFromEnv("X_SECRET"); got != "direct" {
		t.Errorf("got %q, want the variable to win over the file", got)
	}
	t.Setenv("X_SECRET", "")
	if got := secretFromEnv("X_SECRET"); got != "from-file" {
		t.Errorf("got %q, want the trimmed file content", got)
	}
	t.Setenv("X_SECRET_FILE", filepath.Join(t.TempDir(), "missing"))
	if got := secretFromEnv("X_SECRET"); got != "" {
		t.Errorf("got %q, want empty for an unreadable file", got)
	}
}

func TestChainRunsInOrder(t *testing.T) {
	var order []string
	mark := func(name string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				order = append(order, name)
				return next(c)
			}
		}
	}
	handler := chain(mark("auth"), mark("student"))(func(c echo.Context) error {
		order = append(order, "handler")
		return c.NoContent(http.StatusOK)
	})
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if strings.Join(order, ",") != "auth,student,handler" {
		t.Errorf("order = %v", order)
	}
}
