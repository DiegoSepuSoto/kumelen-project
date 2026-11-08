package server

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestConfigFromEnv(t *testing.T) {
	t.Run("defaults outside production", func(t *testing.T) {
		for _, name := range []string{"ENV", "CORS_ALLOWED_ORIGINS", "RATE_LIMIT_RPS", "TRUST_PROXY_HEADERS", "SWAGGER_ENABLED"} {
			t.Setenv(name, "")
		}

		config := ConfigFromEnv()

		if config.Production() || !config.SwaggerEnabled || config.TrustProxyHeaders || config.RateLimitPerSecond != 20 || len(config.AllowedOrigins) != 0 {
			t.Errorf("config = %+v", config)
		}
	})

	t.Run("production closes Swagger and keeps the explicit overrides", func(t *testing.T) {
		t.Setenv("ENV", "Production")
		t.Setenv("CORS_ALLOWED_ORIGINS", " https://app.example.com, ,https://www.example.com ")
		t.Setenv("RATE_LIMIT_RPS", "5.5")
		t.Setenv("TRUST_PROXY_HEADERS", "true")
		t.Setenv("SWAGGER_ENABLED", "")

		config := ConfigFromEnv()

		if !config.Production() || config.SwaggerEnabled || !config.TrustProxyHeaders || config.RateLimitPerSecond != 5.5 {
			t.Errorf("config = %+v", config)
		}
		if want := []string{"https://app.example.com", "https://www.example.com"}; !reflect.DeepEqual(config.AllowedOrigins, want) {
			t.Errorf("origins = %v, want %v", config.AllowedOrigins, want)
		}
		t.Setenv("SWAGGER_ENABLED", "true")
		if !ConfigFromEnv().SwaggerEnabled {
			t.Error("SWAGGER_ENABLED=true must be honored even in production")
		}
	})

	t.Run("invalid values fall back", func(t *testing.T) {
		t.Setenv("RATE_LIMIT_RPS", "mucho")
		t.Setenv("TRUST_PROXY_HEADERS", "quizas")

		if config := ConfigFromEnv(); config.RateLimitPerSecond != 20 || config.TrustProxyHeaders {
			t.Errorf("config = %+v", config)
		}
	})

	t.Run("zero disables the rate limit", func(t *testing.T) {
		t.Setenv("RATE_LIMIT_RPS", "0")
		if ConfigFromEnv().RateLimitPerSecond != 0 {
			t.Error("RATE_LIMIT_RPS=0 must disable the limit")
		}
	})
}

func baseConfig() Config {
	return Config{Environment: "production", RateLimitPerSecond: 0, BodyLimit: "1M"}
}

func newServer(config Config) (*echo.Echo, *bytes.Buffer) {
	logs := &bytes.Buffer{}
	e := echo.New()
	Apply(e, config, slog.New(slog.NewJSONHandler(logs, nil)))
	ok := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	e.GET("/v1/post/search/:term", ok)
	e.GET("/health", ok)
	e.POST("/v1/echo", func(c echo.Context) error {
		if _, err := c.Request().Body.Read(make([]byte, 1)); err != nil && err.Error() != "EOF" {
			return err
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(c.Request().Body); err != nil {
			return err
		}
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/boom", func(c echo.Context) error { panic("boom") })
	return e, logs
}

func do(e *echo.Echo, method, path string, mutate func(*http.Request)) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	if mutate != nil {
		mutate(request)
	}
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)
	return recorder
}

func TestSecurityHeaders(t *testing.T) {
	behindProxy := baseConfig()
	behindProxy.TrustProxyHeaders = true
	production, _ := newServer(behindProxy)

	recorder := do(production, http.MethodGet, "/health", func(r *http.Request) {
		r.RemoteAddr = "10.0.0.5:1234" // proxy privado que termina el TLS
		r.Header.Set("X-Forwarded-Proto", "https")
	})

	header := recorder.Header()
	if header.Get("X-Content-Type-Options") != "nosniff" || header.Get("X-Frame-Options") != "DENY" || header.Get("Referrer-Policy") != "no-referrer" {
		t.Errorf("headers = %v", header)
	}
	if !strings.Contains(header.Get("Content-Security-Policy"), "default-src 'none'") {
		t.Errorf("CSP = %q", header.Get("Content-Security-Policy"))
	}
	if !strings.Contains(header.Get("Strict-Transport-Security"), "max-age=31536000") {
		t.Errorf("HSTS = %q", header.Get("Strict-Transport-Security"))
	}

	t.Run("a client cannot make the server believe the connection is HTTPS", func(t *testing.T) {
		direct, _ := newServer(baseConfig())

		recorder := do(direct, http.MethodGet, "/health", func(r *http.Request) { r.Header.Set("X-Forwarded-Proto", "https") })

		if hsts := recorder.Header().Get("Strict-Transport-Security"); hsts != "" {
			t.Errorf("HSTS = %q, want none for a spoofed header", hsts)
		}
	})

	t.Run("Swagger enabled leaves out the restrictive CSP", func(t *testing.T) {
		config := baseConfig()
		config.SwaggerEnabled = true
		e, _ := newServer(config)

		if csp := do(e, http.MethodGet, "/health", nil).Header().Get("Content-Security-Policy"); csp != "" {
			t.Errorf("CSP = %q, want none so Swagger UI keeps working", csp)
		}
	})

	t.Run("no HSTS outside production", func(t *testing.T) {
		config := behindProxy
		config.Environment = "local"
		e, _ := newServer(config)

		recorder := do(e, http.MethodGet, "/health", func(r *http.Request) {
			r.RemoteAddr = "10.0.0.5:1234"
			r.Header.Set("X-Forwarded-Proto", "https")
		})
		if hsts := recorder.Header().Get("Strict-Transport-Security"); hsts != "" {
			t.Errorf("HSTS = %q", hsts)
		}
	})
}

func TestCORS(t *testing.T) {
	const allowed, other = "https://app.example.com", "https://evil.example.org"
	withOrigin := func(origin string) func(*http.Request) {
		return func(r *http.Request) { r.Header.Set("Origin", origin) }
	}

	t.Run("only the configured origins", func(t *testing.T) {
		config := baseConfig()
		config.AllowedOrigins = []string{allowed}
		e, _ := newServer(config)

		if got := do(e, http.MethodGet, "/health", withOrigin(allowed)).Header().Get("Access-Control-Allow-Origin"); got != allowed {
			t.Errorf("allowed origin got %q", got)
		}
		if got := do(e, http.MethodGet, "/health", withOrigin(other)).Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("other origin got %q, want none", got)
		}
	})

	t.Run("production without origins allows no web origin", func(t *testing.T) {
		e, _ := newServer(baseConfig())

		if got := do(e, http.MethodGet, "/health", withOrigin(other)).Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("got %q, want none", got)
		}
	})

	t.Run("development without origins allows any", func(t *testing.T) {
		config := baseConfig()
		config.Environment = "local"
		e, _ := newServer(config)

		if got := do(e, http.MethodGet, "/health", withOrigin(other)).Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("got %q, want *", got)
		}
	})
}

func fromIP(ip string) func(*http.Request) {
	return func(r *http.Request) { r.RemoteAddr = ip + ":1234" }
}

func TestRateLimit(t *testing.T) {
	config := baseConfig()
	config.RateLimitPerSecond = 1 // ráfaga de 2

	t.Run("limits per client and exempts the probes", func(t *testing.T) {
		e, _ := newServer(config)
		statuses := []int{}
		for i := 0; i < 4; i++ {
			statuses = append(statuses, do(e, http.MethodGet, "/v1/post/search/beca", fromIP("203.0.113.1")).Code)
		}
		if want := []int{200, 200, 429, 429}; !reflect.DeepEqual(statuses, want) {
			t.Errorf("statuses = %v, want %v", statuses, want)
		}
		if code := do(e, http.MethodGet, "/v1/post/search/beca", fromIP("203.0.113.2")).Code; code != http.StatusOK {
			t.Errorf("another client = %d, want 200", code)
		}
		for i := 0; i < 5; i++ {
			if code := do(e, http.MethodGet, "/health", fromIP("203.0.113.1")).Code; code != http.StatusOK {
				t.Fatalf("health probe = %d, want it exempt from the limit", code)
			}
		}
	})

	t.Run("a spoofed X-Forwarded-For does not evade the limit by default", func(t *testing.T) {
		e, _ := newServer(config)
		var last int
		for i := 0; i < 4; i++ {
			spoofed := "198.51.100." + string(rune('1'+i))
			last = do(e, http.MethodGet, "/v1/post/search/beca", func(r *http.Request) {
				fromIP("203.0.113.9")(r)
				r.Header.Set("X-Forwarded-For", spoofed)
			}).Code
		}
		if last != http.StatusTooManyRequests {
			t.Errorf("last status = %d, want 429 (the header must be ignored)", last)
		}
	})

	t.Run("behind a trusted proxy the forwarded address is used", func(t *testing.T) {
		trusted := config
		trusted.TrustProxyHeaders = true
		e, _ := newServer(trusted)
		viaProxy := func(client string) func(*http.Request) {
			return func(r *http.Request) {
				r.RemoteAddr = "10.0.0.5:1234" // proxy privado
				r.Header.Set("X-Forwarded-For", client)
			}
		}
		for i := 0; i < 3; i++ {
			do(e, http.MethodGet, "/v1/post/search/beca", viaProxy("198.51.100.7"))
		}
		if code := do(e, http.MethodGet, "/v1/post/search/beca", viaProxy("198.51.100.7")).Code; code != http.StatusTooManyRequests {
			t.Errorf("same client = %d, want 429", code)
		}
		if code := do(e, http.MethodGet, "/v1/post/search/beca", viaProxy("198.51.100.8")).Code; code != http.StatusOK {
			t.Errorf("different client = %d, want 200", code)
		}
	})

	t.Run("zero disables it", func(t *testing.T) {
		disabled := config
		disabled.RateLimitPerSecond = 0
		e, _ := newServer(disabled)
		for i := 0; i < 10; i++ {
			if code := do(e, http.MethodGet, "/v1/post/search/beca", fromIP("203.0.113.1")).Code; code != http.StatusOK {
				t.Fatalf("request %d = %d", i, code)
			}
		}
	})
}

func TestBodyLimit(t *testing.T) {
	config := baseConfig()
	config.BodyLimit = "1K"
	e, _ := newServer(config)

	post := func(size int) int {
		request := httptest.NewRequest(http.MethodPost, "/v1/echo", strings.NewReader(strings.Repeat("a", size)))
		recorder := httptest.NewRecorder()
		e.ServeHTTP(recorder, request)
		return recorder.Code
	}

	if code := post(100); code != http.StatusOK {
		t.Errorf("small body = %d", code)
	}
	if code := post(4096); code != http.StatusRequestEntityTooLarge {
		t.Errorf("large body = %d, want 413", code)
	}
}

func TestRequestLogContainsNoPersonalData(t *testing.T) {
	e, logs := newServer(baseConfig())

	do(e, http.MethodGet, "/v1/post/search/mi-tema-privado?email=ana@example.com", func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer secreto-del-token")
	})
	do(e, http.MethodGet, "/health", nil)
	do(e, http.MethodGet, "/no-existe/ana@example.com", nil)

	output := logs.String()
	for _, leaked := range []string{"mi-tema-privado", "ana@example.com", "secreto-del-token", "email="} {
		if strings.Contains(output, leaked) {
			t.Errorf("the log leaks %q:\n%s", leaked, output)
		}
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("log lines = %d, want 2 (the health probe is not logged):\n%s", len(lines), output)
	}
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatal(err)
	}
	if entry["route"] != "/v1/post/search/:term" || entry["status"] != float64(200) || entry["method"] != "GET" || entry["request_id"] == "" {
		t.Errorf("entry = %v", entry)
	}
	if !strings.Contains(lines[1], `"route":"unmatched"`) || !strings.Contains(lines[1], `"level":"WARN"`) {
		t.Errorf("unmatched route entry = %s", lines[1])
	}
}

func TestPanicsBecomeServerErrors(t *testing.T) {
	e, logs := newServer(baseConfig())

	if code := do(e, http.MethodGet, "/boom", nil).Code; code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", code)
	}
	if output := logs.String(); !strings.Contains(output, `"level":"ERROR"`) || !strings.Contains(output, "panic recovered") {
		t.Errorf("a panic must be logged as a structured error:\n%s", output)
	}
}
