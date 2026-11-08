package hooks

import (
	"crypto/subtle"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"iluvatar/src/application/usecase"
	"iluvatar/src/domain/models/requests"
	"net/http"
	"os"
)

const (
	postModel      = "post"
	publishEvent   = "entry.publish"
	secretHeader   = "X-Webhook-Secret"
	secretEnvVar   = "WEBHOOK_SECRET"
	basePath       = "/v1/hooks"
	ignoredMessage = "the event was ignored"
)

type hooksHandler struct {
	pushNotificationUseCase usecase.PushNotificationUseCase
}

func NewHooksHandler(e *echo.Echo, pushNotificationUseCase usecase.PushNotificationUseCase) *hooksHandler {
	h := &hooksHandler{pushNotificationUseCase: pushNotificationUseCase}

	hooksGroup := e.Group(basePath)
	hooksGroup.POST("/post", h.newPostHookHandler, validateWebhookSecret)

	return h
}

// validateWebhookSecret exige que el CMS envíe el secreto compartido (X-Webhook-Secret).
// Si el servicio no tiene WEBHOOK_SECRET configurado rechaza todo: no hay forma segura de aceptar.
func validateWebhookSecret(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		expected := os.Getenv(secretEnvVar)
		if expected == "" {
			return c.JSON(http.StatusServiceUnavailable, echo.Map{"message": "webhook secret is not configured"})
		}
		received := c.Request().Header.Get(secretHeader)
		if subtle.ConstantTimeCompare([]byte(received), []byte(expected)) != 1 {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid webhook secret"})
		}
		return next(c)
	}
}

func (h *hooksHandler) newPostHookHandler(c echo.Context) error {
	var hookRequest *requests.HookRequest

	if err := c.Bind(&hookRequest); err != nil || hookRequest == nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "invalid webhook body"})
	}

	if !isPublishedPost(hookRequest) {
		return c.JSON(http.StatusOK, echo.Map{"message": ignoredMessage})
	}

	marshalledHookEntry, err := json.Marshal(hookRequest.Entry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}

	var newPostHookRequest *requests.HookPostModelRequest
	if errUnmarshal := json.Unmarshal(marshalledHookEntry, &newPostHookRequest); errUnmarshal != nil || newPostHookRequest == nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "the request is not a post model"})
	}

	if err = h.pushNotificationUseCase.NotifyPostPublished(c.Request().Context(), newPostHookRequest); err != nil {
		return c.JSON(http.StatusBadGateway, echo.Map{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "the notification was sent successfully"})
}

// isPublishedPost indica si el evento es la publicación de un post; el CMS avisa por cada contenido
// que se publica (servicios, personas...) y solo los posts generan notificación.
func isPublishedPost(hookRequest *requests.HookRequest) bool {
	return hookRequest.Event == publishEvent && hookRequest.Model == postModel
}
