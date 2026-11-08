package strapi

import (
	"errors"
	"iluvatar/src/domain/models"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// Los fixtures de testdata son respuestas reales de Strapi 5.61 (ver record_test.go).
const (
	paeID    = "epc4gw4lp0czni7ri325yp1e"
	sesaesID = "xt10rnjba1ofyepgd18r8b3u"
)

func TestGetServices(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/servicios": serveFixture(t, "services")})

	services, err := NewServiceRepository(cms.client()).GetServices()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(services) != 3 {
		t.Fatalf("services = %d, want 3", len(services))
	}
	first := services[0]
	if first.ID != paeID || first.Name != "Plan de Apoyo Estudiantil" || !strings.HasPrefix(first.LogoURL, cms.server.URL+"/uploads/logo_pae") {
		t.Errorf("first service = %+v", first)
	}
	cms.onlyCall(t, "/api/servicios").expectQuery(t, map[string]string{
		"fields[0]":                 "Nombre",
		"populate[Logo][fields][0]": "url",
		"sort[0]":                   "Nombre:asc",
		"pagination[limit]":         "100",
	})
}

func TestGetDetailedService(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/servicios/" + paeID: serveFixture(t, "service_detail")})

	page, err := NewServiceRepository(cms.client()).GetDetailedService(paeID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != paeID || page.Name != "Plan de Apoyo Estudiantil" || !strings.Contains(page.Description, "formación integral") {
		t.Errorf("page = %+v", page)
	}
	if !strings.HasSuffix(page.LogoURL, "/uploads/logo_pae_2202f89f53.png") {
		t.Errorf("logo = %q", page.LogoURL)
	}
	wantPersons := []*models.Person{{ID: "wz5qmcng71d88a3l3huc94b2", Name: "Profesional PAE", Charge: "Coordinación", Email: "pae@example.com"}}
	if !reflect.DeepEqual(page.Persons, wantPersons) {
		t.Errorf("persons = %+v, want %+v (a person without photo has an empty picture URL)", page.Persons, wantPersons)
	}
	if len(page.Contacts) != 1 || page.Contacts[0].Type != "Web" || page.Contacts[0].Link != "https://example.com/pae" {
		t.Errorf("contacts = %+v", page.Contacts)
	}
	cms.onlyCall(t, "/api/servicios/"+paeID).expectQuery(t, map[string]string{
		"populate[Logo][fields][0]":                       "url",
		"populate[personas][populate][Imagen][fields][0]": "url",
		"populate[contactos][fields][2]":                  "Link",
	})
}

func TestGetDetailedServiceNotFoundAndEscaping(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{
		"/api/servicios/null":  serveJSON(`{"data":null}`),
		"/api/servicios/a%2Fb": serveStatus(http.StatusNotFound),
	})
	repository := NewServiceRepository(cms.client())

	if _, err := repository.GetDetailedService("null"); !errors.Is(err, ErrNotFound) {
		t.Errorf("data null: error = %v, want ErrNotFound", err)
	}
	if _, err := repository.GetDetailedService("a/b"); !errors.Is(err, ErrNotFound) {
		t.Errorf("404: error = %v, want ErrNotFound", err)
	}
	if len(cms.callsTo("/api/servicios/a%2Fb")) != 1 {
		t.Error("an id with a slash must be escaped so it can't reach another CMS route")
	}
}

func TestGetServiceWithPosts(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/servicios/" + paeID: serveFixture(t, "service_posts")})

	page, err := NewServiceRepository(cms.client()).GetServiceWithPosts(paeID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != paeID || page.Name != "Plan de Apoyo Estudiantil" || len(page.PostsForPreview) != 6 {
		t.Fatalf("page = %+v (posts: %d)", page, len(page.PostsForPreview))
	}
	if first := page.PostsForPreview[0]; first.Title != "PAE: publicación 6" || !strings.HasPrefix(first.ImageURL, cms.server.URL+"/uploads/") {
		t.Errorf("first post = %+v, want the newest one first", first)
	}
	cms.onlyCall(t, "/api/servicios/"+paeID).expectQuery(t, map[string]string{
		"populate[posts][sort][0]": "createdAt:desc",
	})
}

func TestGetServicesWithPosts(t *testing.T) {
	postsByService := map[string]string{
		paeID: `{"data":[
			{"documentId":"p2","Titulo":"Segundo","Resumen":"r2","Imagen":{"url":"/uploads/2.png"}},
			{"documentId":"p1","Titulo":"Primero","Resumen":"r1","Imagen":null}]}`,
	}
	cms := newFakeCMS(t, map[string]http.HandlerFunc{
		"/api/servicios": serveFixture(t, "home"),
		"/api/posts": func(w http.ResponseWriter, r *http.Request) {
			body, ok := postsByService[r.URL.Query().Get("filters[servicio][documentId][$eq]")]
			if !ok {
				body = `{"data":[]}`
			}
			serveJSON(body)(w, r)
		},
	})

	home, err := NewServiceRepository(cms.client()).GetServicesWithPosts()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(home.ServicesWithPosts) != 1 {
		t.Fatalf("services with posts = %d, want 1 (services without posts are left out)", len(home.ServicesWithPosts))
	}
	service := home.ServicesWithPosts[0]
	if service.ID != paeID || service.Abbreviation != "PAE" || len(service.PostsForPreview) != 2 {
		t.Fatalf("service = %+v", service)
	}
	if first := service.PostsForPreview[0]; first.ID != "p2" || first.ImageURL != cms.server.URL+"/uploads/2.png" {
		t.Errorf("first post = %+v", first)
	}
	if second := service.PostsForPreview[1]; second.ImageURL != "" {
		t.Errorf("a post without image has an empty URL, got %q", second.ImageURL)
	}

	postCalls := cms.callsTo("/api/posts")
	if len(postCalls) != 3 {
		t.Fatalf("post queries = %d, want one per service (3)", len(postCalls))
	}
	for _, postCall := range postCalls {
		postCall.expectQuery(t, map[string]string{"pagination[limit]": "5", "sort[0]": "createdAt:desc"})
	}
}

func TestGetServicesWithPostsFailsIfAnyServiceQueryFails(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{
		"/api/servicios": serveFixture(t, "home"),
		"/api/posts": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("filters[servicio][documentId][$eq]") == sesaesID {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			serveJSON(`{"data":[]}`)(w, r)
		},
	})

	if _, err := NewServiceRepository(cms.client()).GetServicesWithPosts(); err == nil {
		t.Fatal("expected an error when a service's posts can't be loaded")
	}
}

func TestGetDetailedPost(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/posts/azovonh0wxutfczvohf9gtv6": serveFixture(t, "post_detail")})

	page, err := NewPostRepository(cms.client()).GetDetailedPost("azovonh0wxutfczvohf9gtv6")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "azovonh0wxutfczvohf9gtv6" || page.Title != "SESAES: publicación 1" || !strings.Contains(page.Body, "Contenido de prueba") {
		t.Errorf("page = %+v", page)
	}
	if page.ServiceID != sesaesID || page.ServiceName != "Servicio de Salud Estudiantil" {
		t.Errorf("service = (%s, %s)", page.ServiceID, page.ServiceName)
	}
	if len(page.Categories) != 1 || page.Categories[0].Name != "Salud" {
		t.Errorf("categories = %+v", page.Categories)
	}
	if !strings.HasPrefix(page.Image, cms.server.URL+"/uploads/post_1_1") {
		t.Errorf("image = %q", page.Image)
	}
}

func TestGetDetailedPostWithoutServiceNorImage(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{
		"/api/posts/x": serveJSON(`{"data":{"documentId":"x","Titulo":"Sin servicio","Cuerpo":"b","Imagen":null,"categorias":[],"servicio":null}}`),
	})

	page, err := NewPostRepository(cms.client()).GetDetailedPost("x")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Image != "" || page.ServiceID != "" || page.ServiceName != "" || page.Categories == nil || len(page.Categories) != 0 {
		t.Errorf("page = %+v", page)
	}
}

func TestGetDetailedPostNotFound(t *testing.T) {
	cms := newFakeCMS(t, nil)

	if _, err := NewPostRepository(cms.client()).GetDetailedPost("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestGetPostsByTerm(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/posts": serveFixture(t, "post_search")})

	result, err := NewPostRepository(cms.client()).GetPostsByTerm("publicación 1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Posts) != 2 || result.Posts[0].Title == "" || result.Posts[0].Summary == "" {
		t.Errorf("result = %+v", result.Posts)
	}
	cms.onlyCall(t, "/api/posts").expectQuery(t, map[string]string{
		"filters[Cuerpo][$containsi]": "publicación 1",
		"sort[0]":                     "createdAt:desc",
		"pagination[limit]":           "100",
	})
}

func TestGetPostsByTermWithoutResultsReturnsAnEmptyList(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/posts": serveJSON(`{"data":[]}`)})

	result, err := NewPostRepository(cms.client()).GetPostsByTerm("nada")

	if err != nil || result.Posts == nil || len(result.Posts) != 0 {
		t.Fatalf("got (%+v, %v), want an empty, non-nil list", result, err)
	}
}

func TestGetCategories(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/categorias": serveFixture(t, "categories")})

	categories, err := NewCategoryRepository(cms.client()).GetCategories()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var names []string
	for _, category := range categories {
		names = append(names, category.Name)
	}
	if !reflect.DeepEqual(names, []string{"Beneficios", "Deporte", "Salud"}) {
		t.Errorf("names = %v", names)
	}
	cms.onlyCall(t, "/api/categorias").expectQuery(t, map[string]string{"sort[0]": "Titulo:asc", "fields[0]": "Titulo"})
}

func TestGetCategoryWithPosts(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/categorias/rd7fq5xjrmavkwi7erfb0lqu": serveFixture(t, "category_posts")})

	page, err := NewCategoryRepository(cms.client()).GetCategoryWithPosts("rd7fq5xjrmavkwi7erfb0lqu")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "rd7fq5xjrmavkwi7erfb0lqu" || page.Name != "Salud" || len(page.PostsForPreview) != 7 {
		t.Errorf("page = %+v (posts: %d)", page, len(page.PostsForPreview))
	}
	cms.onlyCall(t, "/api/categorias/rd7fq5xjrmavkwi7erfb0lqu").expectQuery(t, map[string]string{"populate[posts][sort][0]": "createdAt:desc"})
}

func TestGetNotifications(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{"/api/avisos": serveFixture(t, "notifications")})

	page, err := NewNotificationRepository(cms.client()).GetNotifications()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Notifications) != 3 {
		t.Fatalf("notifications = %d, want 3", len(page.Notifications))
	}
	first := page.Notifications[0]
	if first.Title != "Aviso de Deportes" || first.Body != "Texto del aviso." || first.ServiceName != "Servicio de Deportes" ||
		!strings.HasPrefix(first.ServiceImageURL, cms.server.URL+"/uploads/logo_deportes") {
		t.Errorf("first notification = %+v", first)
	}
	cms.onlyCall(t, "/api/avisos").expectQuery(t, map[string]string{"sort[0]": "createdAt:desc", "pagination[limit]": "10"})
}

func TestGetNotificationsWithoutService(t *testing.T) {
	cms := newFakeCMS(t, map[string]http.HandlerFunc{
		"/api/avisos": serveJSON(`{"data":[{"documentId":"a1","Titulo":"Huérfano","Cuerpo":"b","servicio":null}]}`),
	})

	page, err := NewNotificationRepository(cms.client()).GetNotifications()

	if err != nil || len(page.Notifications) != 1 {
		t.Fatalf("got (%+v, %v)", page, err)
	}
	if n := page.Notifications[0]; n.ServiceName != "" || n.ServiceImageURL != "" {
		t.Errorf("notification = %+v", n)
	}
}
