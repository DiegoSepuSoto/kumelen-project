package strapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (f *fakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

// cachedCMS es un CMS falso que cuenta las consultas y puede caerse.
type cachedCMS struct {
	hits   atomic.Int32
	status atomic.Int32 // estado a responder; 0 = 200
	body   atomic.Value
	delay  time.Duration
	client *Client
	clock  *fakeClock
	server *httptest.Server
}

func newCachedCMS(t *testing.T, ttl, stale time.Duration, maxEntries int) *cachedCMS {
	t.Helper()
	cms := &cachedCMS{clock: &fakeClock{now: time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)}}
	cms.body.Store(`{"data":[{"documentId":"a","Titulo":"Primero"}]}`)
	cms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cms.hits.Add(1)
		time.Sleep(cms.delay)
		if status := int(cms.status.Load()); status != 0 {
			w.WriteHeader(status)
			return
		}
		_, _ = w.Write([]byte(cms.body.Load().(string)))
	}))
	t.Cleanup(cms.server.Close)
	cms.client = NewClient(cms.server.URL, "", cms.server.Client(), WithCache(ttl, stale, maxEntries))
	cms.client.now = cms.clock.Now
	return cms
}

func (c *cachedCMS) fetch(t *testing.T, path string, query url.Values) (string, error) {
	t.Helper()
	var out listResponse[postEntity]
	if err := c.client.get(path, query, &out); err != nil {
		return "", err
	}
	return out.Data[0].Titulo, nil
}

func TestCacheServesFreshResponsesWithoutHittingTheCMS(t *testing.T) {
	cms := newCachedCMS(t, time.Minute, time.Hour, 10)

	for i := 0; i < 3; i++ {
		if title, err := cms.fetch(t, "posts", nil); err != nil || title != "Primero" {
			t.Fatalf("call %d: (%q, %v)", i, title, err)
		}
	}
	if hits := cms.hits.Load(); hits != 1 {
		t.Errorf("CMS hits = %d, want 1", hits)
	}

	cms.clock.Advance(61 * time.Second)
	cms.body.Store(`{"data":[{"documentId":"a","Titulo":"Actualizado"}]}`)
	if title, _ := cms.fetch(t, "posts", nil); title != "Actualizado" {
		t.Errorf("after the TTL the response must be refreshed, got %q", title)
	}
	if hits := cms.hits.Load(); hits != 2 {
		t.Errorf("CMS hits = %d, want 2", hits)
	}
}

func TestCacheKeysIncludeThePathAndTheQuery(t *testing.T) {
	cms := newCachedCMS(t, time.Minute, time.Hour, 10)

	_, _ = cms.fetch(t, "posts", url.Values{"a": {"1"}})
	_, _ = cms.fetch(t, "posts", url.Values{"a": {"2"}})
	_, _ = cms.fetch(t, "servicios", url.Values{"a": {"1"}})
	_, _ = cms.fetch(t, "posts", url.Values{"a": {"1"}})

	if hits := cms.hits.Load(); hits != 3 {
		t.Errorf("CMS hits = %d, want 3 (three distinct requests)", hits)
	}
}

func TestStaleResponseIsServedWhenTheCMSFails(t *testing.T) {
	cms := newCachedCMS(t, time.Minute, time.Hour, 10)
	if _, err := cms.fetch(t, "posts", nil); err != nil {
		t.Fatal(err)
	}

	cms.status.Store(http.StatusServiceUnavailable)
	cms.clock.Advance(10 * time.Minute) // vencida, pero dentro del respaldo

	if title, err := cms.fetch(t, "posts", nil); err != nil || title != "Primero" {
		t.Fatalf("got (%q, %v), want the stale response", title, err)
	}

	cms.clock.Advance(2 * time.Hour) // ya ni como respaldo
	if _, err := cms.fetch(t, "posts", nil); err == nil {
		t.Error("an expired stale response must not be served")
	}
}

func TestFailureWithoutAPreviousResponseIsAnError(t *testing.T) {
	cms := newCachedCMS(t, time.Minute, time.Hour, 10)
	cms.status.Store(http.StatusInternalServerError)

	if _, err := cms.fetch(t, "posts", nil); err == nil {
		t.Fatal("expected an error")
	}
}

func TestNotFoundIsNeverServedFromTheCache(t *testing.T) {
	cms := newCachedCMS(t, time.Minute, time.Hour, 10)
	if _, err := cms.fetch(t, "posts", nil); err != nil {
		t.Fatal(err)
	}

	cms.status.Store(http.StatusNotFound)
	cms.clock.Advance(2 * time.Minute)
	if _, err := cms.fetch(t, "posts", nil); err != ErrNotFound {
		t.Fatalf("error = %v, want ErrNotFound (the document was deleted)", err)
	}

	// Aunque el CMS se caiga después, no se resucita lo que se borró.
	cms.status.Store(http.StatusServiceUnavailable)
	if _, err := cms.fetch(t, "posts", nil); err == nil {
		t.Error("a deleted document must not come back from the stale cache")
	}
}

func TestConcurrentRequestsShareASingleCMSQuery(t *testing.T) {
	cms := newCachedCMS(t, time.Minute, time.Hour, 10)
	cms.delay = 100 * time.Millisecond

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := cms.fetch(t, "posts", nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if hits := cms.hits.Load(); hits != 1 {
		t.Errorf("CMS hits = %d, want 1 for 20 simultaneous requests", hits)
	}
}

func TestCacheSizeIsBounded(t *testing.T) {
	cms := newCachedCMS(t, time.Minute, time.Hour, 2)

	for i, key := range []string{"1", "2", "3"} {
		cms.clock.Advance(time.Second)
		_, _ = cms.fetch(t, "posts", url.Values{"k": {key}})
		if size := len(cms.client.cache); size > 2 {
			t.Fatalf("after %d requests the cache holds %d entries, want at most 2", i+1, size)
		}
	}
	if _, kept := cms.client.cache[cms.server.URL+"/api/posts?k=1"]; kept {
		t.Error("the oldest entry should have been evicted")
	}
}

func TestCacheDisabledAlwaysQueriesTheCMS(t *testing.T) {
	cms := newCachedCMS(t, 0, 0, 0)

	for i := 0; i < 3; i++ {
		_, _ = cms.fetch(t, "posts", nil)
	}
	if hits := cms.hits.Load(); hits != 3 {
		t.Errorf("CMS hits = %d, want 3", hits)
	}
	cms.status.Store(http.StatusServiceUnavailable)
	if _, err := cms.fetch(t, "posts", nil); err == nil {
		t.Error("without a cache a CMS failure is an error")
	}
}

func TestResponseLargerThanTheLimitIsRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(make([]byte, maxResponseBytes+10))
	}))
	defer server.Close()

	var out listResponse[postEntity]
	if err := NewClient(server.URL, "", server.Client()).get("posts", nil, &out); err == nil {
		t.Fatal("expected an error for an oversized response")
	}
}

func TestPing(t *testing.T) {
	status := http.StatusNoContent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_health" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(status)
	}))
	defer server.Close()
	client := NewClient(server.URL, "", server.Client())

	if err := client.Ping(context.Background()); err != nil {
		t.Errorf("healthy CMS: %v", err)
	}
	status = http.StatusBadGateway
	if err := client.Ping(context.Background()); err == nil {
		t.Error("an unhealthy CMS must fail the check")
	}
	server.Close()
	if err := client.Ping(context.Background()); err == nil {
		t.Error("an unreachable CMS must fail the check")
	}
}
