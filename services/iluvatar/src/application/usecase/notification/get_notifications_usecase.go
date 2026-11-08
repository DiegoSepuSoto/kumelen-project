package notification

import "iluvatar/src/domain/models"

func (u *notificationUseCase) GetNotifications() (*models.NotificationsPage, error) {
	return u.notificationRepository.GetNotifications()
}
