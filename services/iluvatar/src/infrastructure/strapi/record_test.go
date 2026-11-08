//go:build record

package strapi

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Graba las respuestas reales del CMS como fixtures: STRAPI_URL=... STRAPI_TOKEN=... go test -tags record -run TestRecord
type recorder struct {
	current string
	count   map[string]int
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(response.Body)
	response.Body = io.NopCloser(bytes.NewReader(body))
	name := r.current
	if r.count[name] > 0 {
		name += "_" + string(rune('0'+r.count[name]))
	}
	r.count[name]++
	r.count[r.current]++
	_ = os.WriteFile(filepath.Join("testdata", name+".json"), body, 0o644)
	return response, nil
}

func TestRecord(t *testing.T) {
	url, token := os.Getenv("STRAPI_URL"), os.Getenv("STRAPI_TOKEN")
	rec := &recorder{count: map[string]int{}}
	client := NewClient(url, token, &http.Client{Transport: rec})

	rec.current = "home"
	home, err := NewServiceRepository(client).GetServicesWithPosts()
	if err != nil || len(home.ServicesWithPosts) == 0 {
		t.Fatal(home, err)
	}
	var paeID string
	rec.current = "services"
	services, _ := NewServiceRepository(client).GetServices()
	for _, s := range services {
		if strings.Contains(s.Name, "Apoyo") {
			paeID = s.ID
		}
	}
	rec.current = "service_detail"
	if _, err := NewServiceRepository(client).GetDetailedService(paeID); err != nil {
		t.Fatal(err)
	}
	rec.current = "service_posts"
	if _, err := NewServiceRepository(client).GetServiceWithPosts(paeID); err != nil {
		t.Fatal(err)
	}
	rec.current = "post_search"
	found, err := NewPostRepository(client).GetPostsByTerm("publicación 1")
	if err != nil || len(found.Posts) == 0 {
		t.Fatal(found, err)
	}
	rec.current = "post_detail"
	if _, err := NewPostRepository(client).GetDetailedPost(found.Posts[0].ID); err != nil {
		t.Fatal(err)
	}
	rec.current = "categories"
	categories, err := NewCategoryRepository(client).GetCategories()
	if err != nil || len(categories) == 0 {
		t.Fatal(categories, err)
	}
	rec.current = "category_posts"
	if _, err := NewCategoryRepository(client).GetCategoryWithPosts(categories[len(categories)-1].ID); err != nil {
		t.Fatal(err)
	}
	rec.current = "notifications"
	if _, err := NewNotificationRepository(client).GetNotifications(); err != nil {
		t.Fatal(err)
	}
}
