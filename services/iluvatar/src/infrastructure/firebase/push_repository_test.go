package firebase

import (
	"context"
	"errors"
	"iluvatar/src/domain/models"
	"reflect"
	"strings"
	"testing"

	"firebase.google.com/go/v4/messaging"
)

type fakeMessagingClient struct {
	sent       []*messaging.Message
	sendErr    error
	operations []string
	response   *messaging.TopicManagementResponse
	topicErr   error
}

func (f *fakeMessagingClient) Send(_ context.Context, message *messaging.Message) (string, error) {
	f.sent = append(f.sent, message)
	return "message-id", f.sendErr
}

func (f *fakeMessagingClient) SubscribeToTopic(_ context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	f.operations = append(f.operations, "subscribe "+strings.Join(tokens, ",")+" "+topic)
	return f.response, f.topicErr
}

func (f *fakeMessagingClient) UnsubscribeFromTopic(_ context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	f.operations = append(f.operations, "unsubscribe "+strings.Join(tokens, ",")+" "+topic)
	return f.response, f.topicErr
}

var notification = &models.PushNotification{Title: "Nueva publicación de PAE", Body: "Beca", Data: map[string]string{"post_id": "p1"}}

func TestSendToASingleTopic(t *testing.T) {
	client := &fakeMessagingClient{}

	err := NewPushRepository(client).SendToTopics(context.Background(), notification, []string{"all"})

	if err != nil || len(client.sent) != 1 {
		t.Fatalf("err = %v, messages = %d", err, len(client.sent))
	}
	message := client.sent[0]
	if message.Topic != "all" || message.Condition != "" {
		t.Errorf("topic = %q, condition = %q", message.Topic, message.Condition)
	}
	if message.Notification.Title != notification.Title || message.Notification.Body != "Beca" || !reflect.DeepEqual(message.Data, notification.Data) {
		t.Errorf("message = %+v / %+v", message, message.Notification)
	}
}

func TestSendToSeveralTopicsUsesOneConditionSoNoDeviceGetsItTwice(t *testing.T) {
	client := &fakeMessagingClient{}

	err := NewPushRepository(client).SendToTopics(context.Background(), notification, []string{"all", "service.abc"})

	if err != nil || len(client.sent) != 1 {
		t.Fatalf("err = %v, messages = %d (want exactly one)", err, len(client.sent))
	}
	if got, want := client.sent[0].Condition, "'all' in topics || 'service.abc' in topics"; got != want || client.sent[0].Topic != "" {
		t.Errorf("condition = %q, topic = %q; want %q and no topic", got, client.sent[0].Topic, want)
	}
}

func TestSendErrors(t *testing.T) {
	errFCM := errors.New("fcm down")

	if err := NewPushRepository(&fakeMessagingClient{sendErr: errFCM}).SendToTopics(context.Background(), notification, []string{"all"}); !errors.Is(err, errFCM) {
		t.Errorf("error = %v, want it to wrap %v", err, errFCM)
	}
	if err := NewPushRepository(&fakeMessagingClient{}).SendToTopics(context.Background(), notification, nil); err == nil {
		t.Error("sending without topics must fail")
	}
}

func TestSubscribeAndUnsubscribe(t *testing.T) {
	client := &fakeMessagingClient{response: &messaging.TopicManagementResponse{SuccessCount: 1}}
	repository := NewPushRepository(client)

	if err := repository.Subscribe(context.Background(), "tok", []string{"all", "service.abc"}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := repository.Unsubscribe(context.Background(), "tok", []string{"all"}); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}

	want := []string{"subscribe tok all", "subscribe tok service.abc", "unsubscribe tok all"}
	if !reflect.DeepEqual(client.operations, want) {
		t.Errorf("operations = %v, want %v", client.operations, want)
	}
}

func TestTopicManagementFailures(t *testing.T) {
	errAPI := errors.New("iid down")

	t.Run("rpc error", func(t *testing.T) {
		err := NewPushRepository(&fakeMessagingClient{topicErr: errAPI}).Subscribe(context.Background(), "tok", []string{"all"})
		if !errors.Is(err, errAPI) {
			t.Errorf("error = %v, want it to wrap %v", err, errAPI)
		}
	})

	t.Run("per-token failure reports the reason and stops", func(t *testing.T) {
		client := &fakeMessagingClient{response: &messaging.TopicManagementResponse{
			FailureCount: 1,
			Errors:       []*messaging.ErrorInfo{{Index: 0, Reason: "invalid-argument"}},
		}}

		err := NewPushRepository(client).Subscribe(context.Background(), "tok", []string{"all", "service.abc"})

		if err == nil || !strings.Contains(err.Error(), "invalid-argument") {
			t.Fatalf("error = %v, want the failure reason", err)
		}
		if len(client.operations) != 1 {
			t.Errorf("operations = %v, want it to stop after the first failure", client.operations)
		}
	})

	t.Run("failure without details", func(t *testing.T) {
		client := &fakeMessagingClient{response: &messaging.TopicManagementResponse{FailureCount: 1}}

		if err := NewPushRepository(client).Unsubscribe(context.Background(), "tok", []string{"all"}); err == nil {
			t.Fatal("expected an error")
		}
	})
}
