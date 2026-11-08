package firebase

import (
	"context"
	"iluvatar/src/domain/models"
	"log/slog"
)

// disabledPushRepository permite correr el servicio en desarrollo sin credenciales de FCM:
// no envía nada y lo deja anotado. main.go se niega a usarlo en producción.
type disabledPushRepository struct{}

func NewDisabledPushRepository() *disabledPushRepository {
	return &disabledPushRepository{}
}

func (*disabledPushRepository) SendToTopics(_ context.Context, notification *models.PushNotification, topics []string) error {
	slog.Info("push disabled: notification not sent", "topics", topics, "title", notification.Title)
	return nil
}

func (*disabledPushRepository) Subscribe(_ context.Context, _ string, topics []string) error {
	slog.Info("push disabled: subscription ignored", "topics", topics)
	return nil
}

func (*disabledPushRepository) Unsubscribe(_ context.Context, _ string, topics []string) error {
	slog.Info("push disabled: unsubscription ignored", "topics", topics)
	return nil
}
