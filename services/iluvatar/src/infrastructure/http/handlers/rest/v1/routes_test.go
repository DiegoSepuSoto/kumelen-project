// Contrato REST que consume la app móvil: rutas, protección con JWT y mapeo de errores.
package v1

import (
	"errors"
	"github.com/labstack/echo/v4"
	"iluvatar/src/application/usecase/catalog"
	"iluvatar/src/application/usecase/category"
	"iluvatar/src/application/usecase/home"
	"iluvatar/src/application/usecase/notification"
	"iluvatar/src/application/usecase/post"
	"iluvatar/src/application/usecase/service"
	"iluvatar/src/domain/models"
	catalogHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/catalog"
	categoryHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/category"
	homeHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/home"
	notificationHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/notification"
	postHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/post"
	serviceHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/service"
	"iluvatar/src/shared/testdoubles"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var errCMS = errors.New("cms down")

func allowAll(next echo.HandlerFunc) echo.HandlerFunc { return next }

func denyAll(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error { return c.NoContent(http.StatusUnauthorized) }
}

// newServer registra todos los handlers con repositorios que devuelven datos fijos, o errCMS si fail es true.
func newServer(fail bool, jwt echo.MiddlewareFunc) *echo.Echo {
	result := func(v interface{}) (interface{}, error) {
		if fail {
			return nil, errCMS
		}
		return v, nil
	}

	services := &testdoubles.ServiceRepository{
		GetServicesWithPostsFunc: func() (*models.HomePage, error) {
			if fail {
				return nil, errCMS
			}
			return &models.HomePage{ServicesWithPosts: []*models.ServiceWithPosts{{ID: "1", Abbreviation: "PAE"}}}, nil
		},
		GetDetailedServiceFunc: func(string) (*models.ServicePage, error) {
			v, err := result(&models.ServicePage{})
			if err != nil {
				return nil, err
			}
			return v.(*models.ServicePage), nil
		},
		GetServiceWithPostsFunc: func(string) (*models.ServicePostsPage, error) {
			v, err := result(&models.ServicePostsPage{})
			if err != nil {
				return nil, err
			}
			return v.(*models.ServicePostsPage), nil
		},
		GetServicesFunc: func() ([]*models.Service, error) {
			if fail {
				return nil, errCMS
			}
			return []*models.Service{{ID: "1", Name: "PAE"}}, nil
		},
	}
	categories := &testdoubles.CategoryRepository{
		GetCategoriesFunc: func() ([]*models.Category, error) { return []*models.Category{{ID: "2", Name: "Salud"}}, nil },
		GetCategoryWithPostsFunc: func(string) (*models.CategoryPostsPage, error) {
			v, err := result(&models.CategoryPostsPage{})
			if err != nil {
				return nil, err
			}
			return v.(*models.CategoryPostsPage), nil
		},
	}
	posts := &testdoubles.PostRepository{
		GetDetailedPostFunc: func(string) (*models.PostPage, error) {
			v, err := result(&models.PostPage{})
			if err != nil {
				return nil, err
			}
			return v.(*models.PostPage), nil
		},
		GetPostsByTermFunc: func(string) (*models.PostSearchResult, error) {
			v, err := result(&models.PostSearchResult{})
			if err != nil {
				return nil, err
			}
			return v.(*models.PostSearchResult), nil
		},
	}
	notifications := &testdoubles.NotificationRepository{
		GetNotificationsFunc: func() (*models.NotificationsPage, error) {
			v, err := result(&models.NotificationsPage{})
			if err != nil {
				return nil, err
			}
			return v.(*models.NotificationsPage), nil
		},
	}

	e := echo.New()
	homeHandler.NewHomeHandler(e, home.NewHomeUseCase(services), jwt)
	postHandler.NewPostHandler(e, post.NewPostUseCase(posts), jwt)
	serviceHandler.NewServiceHandler(e, service.NewServiceUseCase(services), jwt)
	catalogHandler.NewCatalogHandler(e, catalog.NewCatalogUseCase(services, categories), jwt)
	categoryHandler.NewCategoryHandler(e, category.NewCategoryUseCase(categories), jwt)
	notificationHandler.NewNotificationHandler(e, notification.NewNotificationUseCase(notifications), jwt)
	return e
}

var routes = []string{
	"/v1/home",
	"/v1/post/1",
	"/v1/post/search/beca",
	"/v1/service/1",
	"/v1/service/1/posts",
	"/v1/catalog",
	"/v1/category/2/posts",
	"/v1/notification",
}

func get(e *echo.Echo, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestRoutesReturnJSONWhenTheCMSResponds(t *testing.T) {
	e := newServer(false, allowAll)
	for _, path := range routes {
		t.Run(path, func(t *testing.T) {
			recorder := get(e, path)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body: %s)", recorder.Code, recorder.Body)
			}
			if !strings.HasPrefix(recorder.Header().Get("Content-Type"), "application/json") {
				t.Errorf("content type = %q", recorder.Header().Get("Content-Type"))
			}
		})
	}
}

func TestRoutesReturnBadGatewayWhenTheCMSFails(t *testing.T) {
	e := newServer(true, allowAll)
	for _, path := range routes {
		t.Run(path, func(t *testing.T) {
			recorder := get(e, path)

			if recorder.Code != http.StatusBadGateway {
				t.Fatalf("status = %d, want 502 (body: %s)", recorder.Code, recorder.Body)
			}
		})
	}
}

func TestRoutesAreProtectedByTheJWTMiddleware(t *testing.T) {
	e := newServer(false, denyAll)
	for _, path := range routes {
		t.Run(path, func(t *testing.T) {
			if got := get(e, path).Code; got != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", got)
			}
		})
	}
}

func TestHomeResponseShape(t *testing.T) {
	recorder := get(newServer(false, allowAll), "/v1/home")

	want := `{"services_with_posts":[{"id":"1","abbreviation":"PAE","posts_for_preview":null}]}`
	if got := strings.TrimSpace(recorder.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}
