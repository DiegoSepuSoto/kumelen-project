package push

import (
	"context"
	"errors"
	"fmt"
	"iluvatar/src/domain/models"
	"regexp"
	"strings"
)

const (
	maxTopicsPerRequest = 20
	maxDeviceTokenSize  = 4096
)

// ErrInvalidSubscription indica una solicitud de suscripción mal formada (token o topics inválidos).
var ErrInvalidSubscription = errors.New("invalid push subscription")

// El documentId de Strapi es alfanumérico; se acota para que nadie cree topics arbitrarios.
var serviceIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

func isValidServiceID(serviceID string) bool {
	return serviceIDPattern.MatchString(serviceID)
}

func (u *pushNotificationUseCase) Subscribe(ctx context.Context, deviceToken string, topics []string) error {
	if err := validateSubscription(deviceToken, topics); err != nil {
		return err
	}
	return u.pushNotificationRepository.Subscribe(ctx, deviceToken, topics)
}

func (u *pushNotificationUseCase) Unsubscribe(ctx context.Context, deviceToken string, topics []string) error {
	if err := validateSubscription(deviceToken, topics); err != nil {
		return err
	}
	return u.pushNotificationRepository.Unsubscribe(ctx, deviceToken, topics)
}

func validateSubscription(deviceToken string, topics []string) error {
	if deviceToken == "" || len(deviceToken) > maxDeviceTokenSize {
		return fmt.Errorf("%w: the device token is empty or too long", ErrInvalidSubscription)
	}
	if len(topics) == 0 || len(topics) > maxTopicsPerRequest {
		return fmt.Errorf("%w: between 1 and %d topics are required", ErrInvalidSubscription, maxTopicsPerRequest)
	}
	for _, topic := range topics {
		if !isValidTopic(topic) {
			return fmt.Errorf("%w: unknown topic %q", ErrInvalidSubscription, topic)
		}
	}
	return nil
}

func isValidTopic(topic string) bool {
	if topic == models.TopicAll {
		return true
	}
	serviceID, isServiceTopic := strings.CutPrefix(topic, models.ServiceTopic(""))
	return isServiceTopic && isValidServiceID(serviceID)
}
