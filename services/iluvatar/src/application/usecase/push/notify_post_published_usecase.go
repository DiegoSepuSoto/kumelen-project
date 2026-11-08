package push

import (
	"context"
	"fmt"
	"iluvatar/src/domain/models"
	"iluvatar/src/domain/models/requests"
)

func (u *pushNotificationUseCase) NotifyPostPublished(ctx context.Context, post *requests.HookPostModelRequest) error {
	notification := &models.PushNotification{
		Title: "Nueva publicación",
		Body:  post.Title,
		Data:  map[string]string{"post_id": post.ID},
	}
	topics := []string{models.TopicAll}

	if service := post.Service; service != nil {
		if service.ShortName != "" {
			notification.Title = fmt.Sprintf("Nueva publicación de %s", service.ShortName)
		}
		if isValidServiceID(service.ID) {
			notification.Data["service_id"] = service.ID
			topics = append(topics, models.ServiceTopic(service.ID))
		}
	}

	return u.pushNotificationRepository.SendToTopics(ctx, notification, topics)
}
