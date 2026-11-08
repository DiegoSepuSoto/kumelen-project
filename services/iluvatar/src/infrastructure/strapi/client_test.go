package strapi

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestClientSendsTheTokenAndTrimsTheHostSlash(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/ping": serveJSON(`{"data":[]}`)})

	var out listResponse[postEntity]
	if err := cms.client().get("ping", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	recorded := cms.onlyCall(t, "/api/ping")
	if recorded.method != http.MethodGet || recorded.auth != "Bearer "+testToken {
		t.Errorf("call = %+v", recorded)
	}
}

func TestClientWithoutTokenSendsNoAuthorization(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/ping": serveJSON(`{"data":[]}`)})
	client := NewClient(cms.server.URL, "", cms.server.Client())

	var out listResponse[postEntity]
	if err := client.get("ping", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cms.onlyCall(t, "/api/ping").auth; got != "" {
		t.Errorf("authorization = %q, want empty", got)
	}
}

func TestClientErrors(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		check   func(error) bool
		want    string
	}{
		{"404 is ErrNotFound", serveStatus(http.StatusNotFound), func(err error) bool { return errors.Is(err, ErrNotFound) }, "ErrNotFound"},
		{"403 reports the status", serveStatus(http.StatusForbidden), func(err error) bool { return err != nil && strings.Contains(err.Error(), "403") }, "status 403"},
		{"500 reports the status", serveStatus(http.StatusInternalServerError), func(err error) bool { return err != nil && strings.Contains(err.Error(), "500") }, "status 500"},
		{"invalid JSON is an error", serveJSON(`nope`), func(err error) bool { return err != nil && strings.Contains(err.Error(), "can't decode") }, "decode error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/things": tt.handler})

			var out listResponse[postEntity]
			if err := cms.client().get("things", nil, &out); !tt.check(err) {
				t.Fatalf("error = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestClientUnreachableServer(t *testing.T) {
	cms := newFakeCMS(t, nil)
	client := cms.client()
	cms.server.Close()

	var out listResponse[postEntity]
	if err := client.get("things", nil, &out); err == nil {
		t.Fatal("expected an error for an unreachable server")
	}
}

func TestMediaURL(t *testing.T) {
	client := NewClient("https://cms.example.com/", "", http.DefaultClient)

	tests := []struct {
		name  string
		media *media
		want  string
	}{
		{"nil media", nil, ""},
		{"empty url", &media{}, ""},
		{"relative url is prefixed with the host", &media{URL: "/uploads/a.png"}, "https://cms.example.com/uploads/a.png"},
		{"absolute https url is kept", &media{URL: "https://cdn.example.com/a.png"}, "https://cdn.example.com/a.png"},
		{"absolute http url is kept", &media{URL: "http://cdn.example.com/a.png"}, "http://cdn.example.com/a.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := client.mediaURL(tt.media); got != tt.want {
				t.Errorf("mediaURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMediaURLUsesThePublicAddressWhenTheCMSIsReachedInternally(t *testing.T) {
	client := NewClient("http://cms:1337", "", http.DefaultClient, WithPublicURL("https://cms.example.com/"))

	if got := client.mediaURL(&media{URL: "/uploads/a.png"}); got != "https://cms.example.com/uploads/a.png" {
		t.Errorf("relative media = %q, want the public address", got)
	}
	if got := client.mediaURL(&media{URL: "https://cdn.example.com/a.png"}); got != "https://cdn.example.com/a.png" {
		t.Errorf("absolute media = %q, must be kept", got)
	}
}
