package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCheckHealth(t *testing.T) {
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
	defer server.Close()

	if err := checkHealth(server.URL, server.Client()); err != nil {
		t.Errorf("healthy: %v", err)
	}
	status = http.StatusServiceUnavailable
	if err := checkHealth(server.URL, server.Client()); err == nil {
		t.Error("a 503 must fail the check")
	}
	server.Close()
	if err := checkHealth(server.URL, server.Client()); err == nil {
		t.Error("an unreachable server must fail the check")
	}
}

func TestRunHealthcheckExitCode(t *testing.T) {
	if code := runHealthcheck("1"); code != 1 {
		t.Errorf("exit code = %d, want 1 when nothing listens", code)
	}
}

func TestCMSToken(t *testing.T) {
	file := t.TempDir() + "/token"
	if err := os.WriteFile(file, []byte("  token-del-archivo\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, token, tokenFile, want string
	}{
		{"nothing configured", "", "", ""},
		{"from the variable", " del-entorno ", "", "del-entorno"},
		{"the variable wins over the file", "del-entorno", file, "del-entorno"},
		{"from the file, trimmed", "", file, "token-del-archivo"},
		{"missing file", "", t.TempDir() + "/no-existe", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ILUVATAR_TOKEN", tt.token)
			t.Setenv("ILUVATAR_TOKEN_FILE", tt.tokenFile)

			if got := cmsToken(); got != tt.want {
				t.Errorf("cmsToken() = %q, want %q", got, tt.want)
			}
		})
	}
}
