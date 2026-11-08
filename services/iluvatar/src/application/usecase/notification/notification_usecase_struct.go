package notification

import "iluvatar/src/domain/repository"

type notificationUseCase struct {
	notificationRepository repository.NotificationRepository
}

func NewNotificationUseCase(notificationRepository repository.NotificationRepository) *notificationUseCase {
	return &notificationUseCase{notificationRepository: notificationRepository}
}
