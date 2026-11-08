package strapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const testToken = "test-token"

// fakeCMS es un servidor que imita al CMS: responde según la ruta (con escapes) y registra las peticiones.
type fakeCMS struct {
	server *httptest.Server
	mu     sync.Mutex
	calls  []call
}

type call struct {
	method string
	path   string
	query  url.Values
	auth   string
}

func newFakeCMS(t *testing.T, routes map[string]http.HandlerFunc) *fakeCMS {
	t.Helper()
	cms := &fakeCMS{}
	cms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cms.mu.Lock()
		cms.calls = append(cms.calls, call{r.Method, r.URL.EscapedPath(), r.URL.Query(), r.Header.Get("Authorization")})
		cms.mu.Unlock()

		handler, ok := routes[r.URL.EscapedPath()]
		if !ok {
			http.NotFound(w, r)
			return
		}
		handler(w, r)
	}))
	t.Cleanup(cms.server.Close)
	return cms
}

func (c *fakeCMS) client() *Client {
	return NewClient(c.server.URL+"/", testToken, c.server.Client())
}

func (c *fakeCMS) callsTo(path string) []call {
	c.mu.Lock()
	defer c.mu.Unlock()
	var matching []call
	for _, recorded := range c.calls {
		if recorded.path == path {
			matching = append(matching, recorded)
		}
	}
	return matching
}

// onlyCall devuelve la única petición hecha a path y falla si hubo otra cantidad.
func (c *fakeCMS) onlyCall(t *testing.T, path string) call {
	t.Helper()
	calls := c.callsTo(path)
	if len(calls) != 1 {
		t.Fatalf("calls to %s = %d, want 1", path, len(calls))
	}
	return calls[0]
}

func (c call) expectQuery(t *testing.T, want map[string]string) {
	t.Helper()
	for key, value := range want {
		if got := c.query.Get(key); got != value {
			t.Errorf("query %s = %q, want %q", key, got, value)
		}
	}
}

func serveFixture(t *testing.T, name string) http.HandlerFunc {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name+".json"))
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return serveJSON(string(body))
}

func serveJSON(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

func serveStatus(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }
}
