package repository

import (
	"context"
	"iluvatar/src/domain/models"
)

type PushNotificationRepository interface {
	// SendToTopics envía la notificación a los dispositivos suscritos a alguno de los topics, sin duplicados.
	SendToTopics(ctx context.Context, notification *models.PushNotification, topics []string) error
	Subscribe(ctx context.Context, deviceToken string, topics []string) error
	Unsubscribe(ctx context.Context, deviceToken string, topics []string) error
}
