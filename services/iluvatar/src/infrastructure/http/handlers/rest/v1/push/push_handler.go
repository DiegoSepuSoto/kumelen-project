package push

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"iluvatar/src/application/usecase"
	pushUseCase "iluvatar/src/application/usecase/push"
	"net/http"
)

const basePath = "/v1/push"

type pushHandler struct {
	pushNotificationUseCase usecase.PushNotificationUseCase
}

type subscriptionRequest struct {
	DeviceToken string   `json:"device_token" validate:"required"`
	Topics      []string `json:"topics" validate:"required,min=1"`
}

// NewPushHandler registra la suscripción de dispositivos a topics de notificaciones. El cliente web
// no puede suscribirse por sí solo a un topic de FCM, así que lo hace el servidor con el token del dispositivo.
func NewPushHandler(e *echo.Echo, pushNotificationUseCase usecase.PushNotificationUseCase, authMiddleware echo.MiddlewareFunc) *pushHandler {
	h := &pushHandler{pushNotificationUseCase: pushNotificationUseCase}

	pushGroup := e.Group(basePath)
	pushGroup.Use(authMiddleware)
	pushGroup.POST("/subscribe", h.subscribe)
	pushGroup.POST("/unsubscribe", h.unsubscribe)

	return h
}

// @Summary Suscripción a notificaciones
// @Description Suscribe el token de un dispositivo a los topics indicados ("all" o "service.<id del servicio>")
// @Tags API V1 - Notificaciones push
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <ID token de Firebase>"
// @Param body body subscriptionRequest true "Token del dispositivo y topics"
// @Success 200 {object} map[string]interface{} "OK"
// @Failure 400 {object} map[string]interface{} "BadRequest"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 502 {object} map[string]interface{} "BadGateway"
// @Router /v1/push/subscribe [post]
func (h *pushHandler) subscribe(c echo.Context) error {
	return h.handle(c, h.pushNotificationUseCase.Subscribe, "subscribed")
}

// @Summary Baja de notificaciones
// @Description Quita al dispositivo de los topics indicados
// @Tags API V1 - Notificaciones push
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <ID token de Firebase>"
// @Param body body subscriptionRequest true "Token del dispositivo y topics"
// @Success 200 {object} map[string]interface{} "OK"
// @Failure 400 {object} map[string]interface{} "BadRequest"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 502 {object} map[string]interface{} "BadGateway"
// @Router /v1/push/unsubscribe [post]
func (h *pushHandler) unsubscribe(c echo.Context) error {
	return h.handle(c, h.pushNotificationUseCase.Unsubscribe, "unsubscribed")
}

type subscriptionOperation func(ctx context.Context, deviceToken string, topics []string) error

func (h *pushHandler) handle(c echo.Context, operation subscriptionOperation, successMessage string) error {
	var request subscriptionRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "invalid request body"})
	}
	if err := c.Validate(&request); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}

	if err := operation(c.Request().Context(), request.DeviceToken, request.Topics); err != nil {
		if errors.Is(err, pushUseCase.ErrInvalidSubscription) {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
		}
		return c.JSON(http.StatusBadGateway, echo.Map{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": successMessage})
}
