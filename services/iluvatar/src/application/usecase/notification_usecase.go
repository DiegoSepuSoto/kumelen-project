package usecase

import "iluvatar/src/domain/models"

type NotificationUseCase interface {
	GetNotifications() (*models.NotificationsPage, error)
}
