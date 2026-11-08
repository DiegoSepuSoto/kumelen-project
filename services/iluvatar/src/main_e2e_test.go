//go:build e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	firebaseAdapter "iluvatar/src/infrastructure/firebase"
	customMiddlewares "iluvatar/src/infrastructure/http/handlers/rest/middleware"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Prueba de punta a punta del BFF contra un Strapi 5 real (datos de `npm run seed`) y el emulador de Firebase Auth:
//
//	firebase emulators:start --only auth --project demo-kumelen
//	FIREBASE_AUTH_EMULATOR_HOST=127.0.0.1:9099 FIREBASE_PROJECT_ID=demo-kumelen \
//	ILUVATAR_CMS_HOST=http://localhost:1337 ILUVATAR_TOKEN=... go test -tags e2e -run TestBFFAgainstRealCMS ./src/
func TestBFFAgainstRealCMS(t *testing.T) {
	host, emulator, projectID := os.Getenv("ILUVATAR_CMS_HOST"), os.Getenv("FIREBASE_AUTH_EMULATOR_HOST"), os.Getenv("FIREBASE_PROJECT_ID")
	if host == "" || os.Getenv("ILUVATAR_TOKEN") == "" || emulator == "" || projectID == "" {
		t.Skip("faltan ILUVATAR_CMS_HOST, ILUVATAR_TOKEN, FIREBASE_AUTH_EMULATOR_HOST o FIREBASE_PROJECT_ID")
	}

	app, err := firebaseAdapter.NewApp(context.Background(), projectID)
	if err != nil {
		t.Fatal(err)
	}
	authClient, err := app.Auth(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	registerBFFHandlers(e, customMiddlewares.RequireAuth(firebaseAdapter.NewAuthVerifier(authClient)), newCMSClient())

	studentToken := signUpInEmulator(t, emulator)

	get := func(path, token string, out interface{}) int {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		if token != "" {
			request.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		e.ServeHTTP(recorder, request)
		if out != nil && recorder.Code == http.StatusOK {
			if err := json.Unmarshal(recorder.Body.Bytes(), out); err != nil {
				t.Fatalf("GET %s: invalid JSON: %v\n%s", path, err, recorder.Body)
			}
		}
		return recorder.Code
	}

	if code := get("/v1/home", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("GET /v1/home without token = %d, want 401", code)
	}

	var home struct {
		Services []struct {
			ID    string `json:"id"`
			Posts []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				ImageURL string `json:"image_url"`
			} `json:"posts_for_preview"`
		} `json:"services_with_posts"`
	}
	if code := get("/v1/home", studentToken, &home); code != http.StatusOK || len(home.Services) == 0 {
		t.Fatalf("GET /v1/home = %d, services = %d", code, len(home.Services))
	}
	for _, service := range home.Services {
		if len(service.Posts) == 0 || len(service.Posts) > 5 {
			t.Errorf("service %s has %d posts, want 1..5", service.ID, len(service.Posts))
		}
		for _, post := range service.Posts {
			if !strings.HasPrefix(post.ImageURL, host+"/uploads/") {
				t.Errorf("post %s image = %q", post.ID, post.ImageURL)
			}
		}
	}
	serviceID, postID := home.Services[0].ID, home.Services[0].Posts[0].ID

	var catalog struct {
		Services   []struct{ ID, Name string } `json:"services"`
		Categories []struct{ ID, Name string } `json:"categories"`
	}
	if code := get("/v1/catalog", studentToken, &catalog); code != http.StatusOK || len(catalog.Services) == 0 || len(catalog.Categories) == 0 {
		t.Fatalf("GET /v1/catalog = %d, %+v", code, catalog)
	}
	categoryID := catalog.Categories[0].ID

	for _, path := range []string{
		"/v1/service/" + serviceID,
		"/v1/service/" + serviceID + "/posts",
		"/v1/post/" + postID,
		"/v1/post/search/publicaci%C3%B3n",
		"/v1/category/" + categoryID + "/posts",
		"/v1/notification",
	} {
		var body map[string]interface{}
		if code := get(path, studentToken, &body); code != http.StatusOK || len(body) == 0 {
			t.Errorf("GET %s = %d, body = %v", path, code, body)
		}
	}

	if code := get("/v1/post/no-existe", studentToken, nil); code != http.StatusBadGateway {
		t.Errorf("GET unknown post = %d, want 502 (the CMS has no such document)", code)
	}
}

// signUpInEmulator crea un usuario en el emulador de Firebase Auth y devuelve su ID token.
func signUpInEmulator(t *testing.T, emulatorHost string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"email": fmt.Sprintf("e2e-%d@example.com", os.Getpid()), "password": "clave-de-prueba-123", "returnSecureToken": true,
	})
	response, err := http.Post("http://"+emulatorHost+"/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var result struct {
		IDToken string `json:"idToken"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil || result.IDToken == "" {
		t.Fatalf("no se pudo crear el usuario en el emulador: estado %d, error %v", response.StatusCode, err)
	}
	return result.IDToken
}
