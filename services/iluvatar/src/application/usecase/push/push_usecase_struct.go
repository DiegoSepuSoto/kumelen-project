package push

import "iluvatar/src/domain/repository"

type pushNotificationUseCase struct {
	pushNotificationRepository repository.PushNotificationRepository
}

func NewPushNotificationUseCase(pushNotificationRepository repository.PushNotificationRepository) *pushNotificationUseCase {
	return &pushNotificationUseCase{pushNotificationRepository: pushNotificationRepository}
}
