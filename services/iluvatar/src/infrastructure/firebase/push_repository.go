package firebase

import (
	"context"
	"errors"
	"fmt"
	"iluvatar/src/domain/models"
	"strings"

	"firebase.google.com/go/v4/messaging"
)

// messagingClient es lo que se usa de *messaging.Client; permite sustituirlo en las pruebas.
type messagingClient interface {
	Send(ctx context.Context, message *messaging.Message) (string, error)
	SubscribeToTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error)
	UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error)
}

type pushRepository struct {
	client messagingClient
}

func NewPushRepository(client messagingClient) *pushRepository {
	return &pushRepository{client: client}
}

// SendToTopics envía un solo mensaje FCM v1. Con varios topics usa una condición (OR), de modo que
// un dispositivo suscrito a más de uno recibe la notificación una sola vez.
func (r *pushRepository) SendToTopics(ctx context.Context, notification *models.PushNotification, topics []string) error {
	if len(topics) == 0 {
		return errors.New("at least one topic is required")
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{Title: notification.Title, Body: notification.Body},
		Data:         notification.Data,
	}
	if len(topics) == 1 {
		message.Topic = topics[0]
	} else {
		conditions := make([]string, 0, len(topics))
		for _, topic := range topics {
			conditions = append(conditions, fmt.Sprintf("'%s' in topics", topic))
		}
		message.Condition = strings.Join(conditions, " || ")
	}

	if _, err := r.client.Send(ctx, message); err != nil {
		return fmt.Errorf("can't send the push notification: %w", err)
	}
	return nil
}

func (r *pushRepository) Subscribe(ctx context.Context, deviceToken string, topics []string) error {
	return r.manageTopics(ctx, deviceToken, topics, r.client.SubscribeToTopic, "subscribe")
}

func (r *pushRepository) Unsubscribe(ctx context.Context, deviceToken string, topics []string) error {
	return r.manageTopics(ctx, deviceToken, topics, r.client.UnsubscribeFromTopic, "unsubscribe")
}

type topicOperation func(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error)

func (r *pushRepository) manageTopics(ctx context.Context, deviceToken string, topics []string, operation topicOperation, action string) error {
	for _, topic := range topics {
		response, err := operation(ctx, []string{deviceToken}, topic)
		if err != nil {
			return fmt.Errorf("can't %s to %s: %w", action, topic, err)
		}
		if response != nil && response.FailureCount > 0 {
			reason := "unknown"
			if len(response.Errors) > 0 {
				reason = response.Errors[0].Reason
			}
			return fmt.Errorf("can't %s to %s: %s", action, topic, reason)
		}
	}
	return nil
}
