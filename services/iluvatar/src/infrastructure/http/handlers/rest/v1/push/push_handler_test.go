package push

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	pushUseCase "iluvatar/src/application/usecase/push"
	"iluvatar/src/domain/models/requests"
	"iluvatar/src/shared/validations"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type fakePushUseCase struct {
	calls []string
	err   error
}

func (f *fakePushUseCase) record(action, token string, topics []string) error {
	f.calls = append(f.calls, fmt.Sprintf("%s|%s|%s", action, token, strings.Join(topics, ",")))
	return f.err
}

func (f *fakePushUseCase) NotifyPostPublished(context.Context, *requests.HookPostModelRequest) error {
	return nil
}

func (f *fakePushUseCase) Subscribe(_ context.Context, token string, topics []string) error {
	return f.record("subscribe", token, topics)
}

func (f *fakePushUseCase) Unsubscribe(_ context.Context, token string, topics []string) error {
	return f.record("unsubscribe", token, topics)
}

func allowAll(next echo.HandlerFunc) echo.HandlerFunc { return next }

func denyAll(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error { return c.NoContent(http.StatusUnauthorized) }
}

func post(useCase *fakePushUseCase, auth echo.MiddlewareFunc, path, body string) *httptest.ResponseRecorder {
	e := echo.New()
	e.Validator = validations.NewCustomValidator(validator.New())
	NewPushHandler(e, useCase, auth)

	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)
	return recorder
}

func TestSubscribeAndUnsubscribe(t *testing.T) {
	body := `{"device_token":"tok-1","topics":["all","service.abc"]}`

	for path, want := range map[string]string{
		"/v1/push/subscribe":   "subscribe|tok-1|all,service.abc",
		"/v1/push/unsubscribe": "unsubscribe|tok-1|all,service.abc",
	} {
		t.Run(path, func(t *testing.T) {
			useCase := &fakePushUseCase{}

			recorder := post(useCase, allowAll, path, body)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body)
			}
			if !reflect.DeepEqual(useCase.calls, []string{want}) {
				t.Errorf("calls = %v, want [%s]", useCase.calls, want)
			}
		})
	}
}

func TestRoutesAreProtected(t *testing.T) {
	for _, path := range []string{"/v1/push/subscribe", "/v1/push/unsubscribe"} {
		useCase := &fakePushUseCase{}

		recorder := post(useCase, denyAll, path, `{"device_token":"tok","topics":["all"]}`)

		if recorder.Code != http.StatusUnauthorized || len(useCase.calls) != 0 {
			t.Errorf("%s: status = %d, calls = %v; want 401 and no calls", path, recorder.Code, useCase.calls)
		}
	}
}

func TestInvalidRequests(t *testing.T) {
	tests := map[string]string{
		"not json":       `nope`,
		"missing token":  `{"topics":["all"]}`,
		"empty token":    `{"device_token":"","topics":["all"]}`,
		"missing topics": `{"device_token":"tok"}`,
		"empty topics":   `{"device_token":"tok","topics":[]}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			useCase := &fakePushUseCase{}

			recorder := post(useCase, allowAll, "/v1/push/subscribe", body)

			if recorder.Code != http.StatusBadRequest || len(useCase.calls) != 0 {
				t.Errorf("status = %d, calls = %v; want 400 and no calls", recorder.Code, useCase.calls)
			}
		})
	}
}

func TestUseCaseErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid subscription is a client error", fmt.Errorf("%w: unknown topic", pushUseCase.ErrInvalidSubscription), http.StatusBadRequest},
		{"firebase failure is a bad gateway", errors.New("can't subscribe to all: UNREGISTERED"), http.StatusBadGateway},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := &fakePushUseCase{err: tt.err}

			recorder := post(useCase, allowAll, "/v1/push/subscribe", `{"device_token":"tok","topics":["all"]}`)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
