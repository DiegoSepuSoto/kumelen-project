package usecase

import (
	"context"
	"iluvatar/src/domain/models/requests"
)

type PushNotificationUseCase interface {
	// NotifyPostPublished avisa a los dispositivos suscritos al servicio del post y al topic general.
	NotifyPostPublished(ctx context.Context, post *requests.HookPostModelRequest) error
	Subscribe(ctx context.Context, deviceToken string, topics []string) error
	Unsubscribe(ctx context.Context, deviceToken string, topics []string) error
}
