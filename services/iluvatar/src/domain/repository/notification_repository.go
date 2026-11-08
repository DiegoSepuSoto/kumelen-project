package repository

import "iluvatar/src/domain/models"

type NotificationRepository interface {
	GetNotifications() (*models.NotificationsPage, error)
}
