package laurelin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeLaurelin struct {
	server   *httptest.Server
	calls    atomic.Int32
	status   atomic.Int32
	response atomic.Value // string
	lastBody atomic.Value
	lastAuth atomic.Value
	lastURL  atomic.Value
	delay    time.Duration
}

func newFake(t *testing.T) *fakeLaurelin {
	t.Helper()
	f := &fakeLaurelin{}
	f.status.Store(http.StatusOK)
	f.response.Store(`{"is_student":true,"status":"active","level":"master","institution_name":"UTEM","program_name":"Datos"}`)
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.calls.Add(1)
		body, _ := io.ReadAll(r.Body)
		f.lastBody.Store(string(body))
		f.lastAuth.Store(r.Header.Get("Authorization"))
		f.lastURL.Store(r.URL.String())
		time.Sleep(f.delay)
		w.WriteHeader(int(f.status.Load()))
		_, _ = io.WriteString(w, f.response.Load().(string))
	}))
	t.Cleanup(f.server.Close)
	return f
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newClient(f *fakeLaurelin, clk *clock, opts ...Option) *Client {
	c := NewClient(f.server.URL+"/", "service-secret", f.server.Client(), opts...)
	c.now = clk.now
	return c
}

var cacheOpts = WithCache(10*time.Minute, time.Minute, 24*time.Hour, 100)

func TestLookupSendsTheTokenAndKeepsTheEmailOutOfTheURL(t *testing.T) {
	f := newFake(t)
	client := newClient(f, &clock{t: time.Unix(0, 0)})

	standing, err := client.Lookup(context.Background(), "Ana@utem.cl")
	if err != nil {
		t.Fatal(err)
	}
	if !standing.IsStudent || standing.Level != "master" || standing.InstitutionName != "UTEM" || standing.ProgramName != "Datos" || standing.Status != "active" {
		t.Errorf("standing = %+v", standing)
	}
	if f.lastAuth.Load() != "Bearer service-secret" {
		t.Errorf("authorization = %v", f.lastAuth.Load())
	}
	if f.lastURL.Load() != "/internal/v1/students/lookup" || strings.Contains(f.lastURL.Load().(string), "utem") {
		t.Errorf("url = %v: the email must not go in the URL", f.lastURL.Load())
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(f.lastBody.Load().(string)), &body); err != nil || body["email"] != "Ana@utem.cl" {
		t.Errorf("body = %v (%v)", f.lastBody.Load(), err)
	}
}

func TestLookupResponses(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantStudent bool
		wantErr     bool
	}{
		{"active student", 200, `{"is_student":true,"status":"active"}`, true, false},
		{"graduated", 200, `{"is_student":false,"status":"graduated"}`, false, false},
		{"unknown email", 404, `{"message":"not found"}`, false, false},
		{"wrong service token", 401, `{"message":"invalid service token"}`, false, true},
		{"service token not configured", 503, `{}`, false, true},
		{"server error", 500, `oops`, false, true},
		{"invalid email", 400, `{}`, false, true},
		{"unreadable body", 200, `not json`, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFake(t)
			f.status.Store(int32(tt.status))
			f.response.Store(tt.body)

			standing, err := newClient(f, &clock{t: time.Unix(0, 0)}).Lookup(context.Background(), "a@utem.cl")

			if (err != nil) != tt.wantErr || (err == nil && standing.IsStudent != tt.wantStudent) {
				t.Errorf("standing = %+v, err = %v", standing, err)
			}
			if err != nil && strings.Contains(err.Error(), "a@utem.cl") {
				t.Errorf("the error leaks the email: %v", err)
			}
		})
	}
}

func TestLookupIsCached(t *testing.T) {
	f := newFake(t)
	clk := &clock{t: time.Unix(0, 0)}
	client := newClient(f, clk, cacheOpts)
	lookup := func(email string) bool {
		standing, err := client.Lookup(context.Background(), email)
		if err != nil {
			t.Fatal(err)
		}
		return standing.IsStudent
	}

	lookup("a@utem.cl")
	lookup("  A@UTEM.cl ") // mismo correo
	if f.calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1 (cached, case-insensitive)", f.calls.Load())
	}
	lookup("b@utem.cl")
	if f.calls.Load() != 2 {
		t.Errorf("calls = %d, want 2 (another email)", f.calls.Load())
	}

	clk.advance(9 * time.Minute)
	lookup("a@utem.cl")
	if f.calls.Load() != 2 {
		t.Errorf("calls = %d: still fresh after 9 minutes", f.calls.Load())
	}
	clk.advance(2 * time.Minute)
	lookup("a@utem.cl")
	if f.calls.Load() != 3 {
		t.Errorf("calls = %d, want 3 after the student TTL", f.calls.Load())
	}
}

func TestNotStudentExpiresQuickly(t *testing.T) {
	f := newFake(t)
	f.status.Store(404)
	clk := &clock{t: time.Unix(0, 0)}
	client := newClient(f, clk, cacheOpts)

	client.Lookup(context.Background(), "nuevo@utem.cl")
	client.Lookup(context.Background(), "nuevo@utem.cl")
	if f.calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", f.calls.Load())
	}

	// Se da de alta al estudiante: a los 61 segundos Laurelin se vuelve a consultar y ya entra.
	f.status.Store(200)
	clk.advance(61 * time.Second)
	standing, err := client.Lookup(context.Background(), "nuevo@utem.cl")
	if err != nil || !standing.IsStudent {
		t.Errorf("standing = %+v, err = %v; want the new student accepted", standing, err)
	}
}

func TestStaleAnswerWhenLaurelinIsDown(t *testing.T) {
	f := newFake(t)
	clk := &clock{t: time.Unix(0, 0)}
	client := newClient(f, clk, cacheOpts)
	client.Lookup(context.Background(), "ana@utem.cl")
	client.Lookup(context.Background(), "ausente@utem.cl")

	f.server.Close() // Laurelin cae
	clk.advance(time.Hour)

	standing, err := client.Lookup(context.Background(), "ana@utem.cl")
	if err != nil || !standing.IsStudent {
		t.Errorf("verified student during the outage: %+v, %v; want accepted from the last answer", standing, err)
	}
	if _, err := client.Lookup(context.Background(), "desconocido@utem.cl"); err == nil {
		t.Error("an email never verified must fail while Laurelin is down (fail closed)")
	}

	clk.advance(25 * time.Hour)
	if _, err := client.Lookup(context.Background(), "ana@utem.cl"); err == nil {
		t.Error("the last answer must expire after 24 hours")
	}
}

func TestStaleAnswerIsNotUsedForNotStudents(t *testing.T) {
	f := newFake(t)
	f.status.Store(404)
	clk := &clock{t: time.Unix(0, 0)}
	client := newClient(f, clk, cacheOpts)
	client.Lookup(context.Background(), "otro@utem.cl")

	f.server.Close()
	clk.advance(5 * time.Minute)
	if standing, err := client.Lookup(context.Background(), "otro@utem.cl"); err == nil {
		t.Errorf("standing = %+v; a not-student must not be accepted or cached as an answer while the service is down", standing)
	}
}

func TestOneRequestForConcurrentLookups(t *testing.T) {
	f := newFake(t)
	f.delay = 100 * time.Millisecond
	client := newClient(f, &clock{t: time.Unix(0, 0)}, cacheOpts)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if standing, err := client.Lookup(context.Background(), "ana@utem.cl"); err != nil || !standing.IsStudent {
				t.Errorf("standing = %+v, err = %v", standing, err)
			}
		}()
	}
	wg.Wait()
	if f.calls.Load() != 1 {
		t.Errorf("calls = %d, want 1 for 20 concurrent lookups", f.calls.Load())
	}
}

func TestCacheIsBounded(t *testing.T) {
	f := newFake(t)
	client := newClient(f, &clock{t: time.Unix(0, 0)}, WithCache(time.Hour, time.Hour, time.Hour, 5))
	for i := 0; i < 50; i++ {
		client.Lookup(context.Background(), "u"+string(rune('a'+i%26))+string(rune('a'+i/26))+"@utem.cl")
	}
	client.mu.Lock()
	defer client.mu.Unlock()
	if len(client.entries) > 5 {
		t.Errorf("entries = %d, want at most 5", len(client.entries))
	}
	for k := range client.entries {
		if strings.Contains(k, "@") {
			t.Errorf("the cache keeps a plain email: %q", k)
		}
	}
}

func TestLookupSurvivesTheCallerGivingUp(t *testing.T) {
	f := newFake(t)
	f.delay = 50 * time.Millisecond
	client := newClient(f, &clock{t: time.Unix(0, 0)}, cacheOpts)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.Lookup(ctx, "ana@utem.cl"); err != nil {
		t.Errorf("a cancelled caller context must not fail the shared lookup: %v", err)
	}
}
