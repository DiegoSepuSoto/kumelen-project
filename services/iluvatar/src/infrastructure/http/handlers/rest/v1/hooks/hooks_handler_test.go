package hooks

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"iluvatar/src/domain/models/requests"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const testSecret = "test-secret"

type fakePushUseCase struct {
	received []*requests.HookPostModelRequest
	err      error
}

func (f *fakePushUseCase) NotifyPostPublished(_ context.Context, request *requests.HookPostModelRequest) error {
	f.received = append(f.received, request)
	return f.err
}

func (f *fakePushUseCase) Subscribe(context.Context, string, []string) error   { return nil }
func (f *fakePushUseCase) Unsubscribe(context.Context, string, []string) error { return nil }

// realPayload es el webhook que Strapi 5.61 envía al publicar un post (ver testdata).
func realPayload(t *testing.T) string {
	t.Helper()
	body, err := os.ReadFile("testdata/strapi_entry_publish_post.json")
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func post(t *testing.T, useCase *fakePushUseCase, secret, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	NewHooksHandler(e, useCase)

	request := httptest.NewRequest(http.MethodPost, "/v1/hooks/post", strings.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if secret != "" {
		request.Header.Set("X-Webhook-Secret", secret)
	}
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)
	return recorder
}

func TestPublishedPostSendsTheNotification(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", testSecret)
	useCase := &fakePushUseCase{}

	recorder := post(t, useCase, testSecret, realPayload(t))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body)
	}
	if len(useCase.received) != 1 {
		t.Fatalf("notifications sent = %d, want 1", len(useCase.received))
	}
	got := useCase.received[0]
	if got.ID != "rxxyxbq7z7pdefevbm95sllf" || !strings.HasPrefix(got.Title, "Prueba de webhook") || got.Overview != "Resumen para webhook" {
		t.Errorf("request = %+v", got)
	}
	if got.Service == nil || got.Service.ShortName != "PAE" || got.Service.Name != "Plan de Apoyo Estudiantil" || got.Service.ID != "epc4gw4lp0czni7ri325yp1e" {
		t.Errorf("service = %+v", got.Service)
	}
}

func TestWebhookSecret(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		sent       string
		wantStatus int
	}{
		{"missing secret is unauthorized", testSecret, "", http.StatusUnauthorized},
		{"wrong secret is unauthorized", testSecret, "other", http.StatusUnauthorized},
		{"unconfigured service rejects even a matching empty secret", "", "", http.StatusServiceUnavailable},
		{"unconfigured service rejects any secret", "", testSecret, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WEBHOOK_SECRET", tt.configured)
			useCase := &fakePushUseCase{}

			recorder := post(t, useCase, tt.sent, realPayload(t))

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if len(useCase.received) != 0 {
				t.Error("no notification may be sent without a valid secret")
			}
		})
	}
}

func TestOtherEventsAndModelsAreIgnored(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", testSecret)

	tests := map[string]string{
		"another model published":     strings.Replace(realPayload(t), `"model": "post"`, `"model": "servicio"`, 1),
		"post updated, not published": strings.Replace(realPayload(t), `"entry.publish"`, `"entry.update"`, 1),
		"post unpublished":            strings.Replace(realPayload(t), `"entry.publish"`, `"entry.unpublish"`, 1),
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			useCase := &fakePushUseCase{}

			recorder := post(t, useCase, testSecret, body)

			if recorder.Code != http.StatusOK || len(useCase.received) != 0 {
				t.Errorf("status = %d, sent = %d; want 200 and no notification", recorder.Code, len(useCase.received))
			}
		})
	}
}

func TestInvalidBodies(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", testSecret)

	for name, body := range map[string]string{"not json": "nope", "null": "null"} {
		t.Run(name, func(t *testing.T) {
			useCase := &fakePushUseCase{}

			if recorder := post(t, useCase, testSecret, body); recorder.Code != http.StatusBadRequest || len(useCase.received) != 0 {
				t.Errorf("status = %d, sent = %d; want 400 and no notification", recorder.Code, len(useCase.received))
			}
		})
	}
}

func TestUseCaseErrorIsReportedAsBadGateway(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", testSecret)
	useCase := &fakePushUseCase{err: errors.New("can't send the push notification: fcm down")}

	recorder := post(t, useCase, testSecret, realPayload(t))

	if recorder.Code != http.StatusBadGateway || !strings.Contains(recorder.Body.String(), "fcm down") {
		t.Errorf("status = %d, body = %s", recorder.Code, recorder.Body)
	}
}
