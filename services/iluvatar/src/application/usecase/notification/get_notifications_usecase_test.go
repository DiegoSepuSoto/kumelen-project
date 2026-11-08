package notification

import (
	"errors"
	"iluvatar/src/domain/models"
	"iluvatar/src/shared/testdoubles"
	"testing"
)

func TestGetNotifications(t *testing.T) {
	page := &models.NotificationsPage{Notifications: []*models.Notification{{ID: "1", Title: "Nueva beca"}}}
	errRepo := errors.New("cms down")

	tests := []struct {
		name    string
		repoErr error
		want    *models.NotificationsPage
	}{
		{name: "returns the notifications page", want: page},
		{name: "propagates the repository error", repoErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewNotificationUseCase(&testdoubles.NotificationRepository{
				GetNotificationsFunc: func() (*models.NotificationsPage, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return page, nil
				},
			})

			got, err := useCase.GetNotifications()

			if !errors.Is(err, tt.repoErr) || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.repoErr)
			}
		})
	}
}
