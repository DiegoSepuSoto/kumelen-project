package push

import (
	"context"
	"errors"
	"iluvatar/src/domain/models"
	"iluvatar/src/domain/models/requests"
	"reflect"
	"strings"
	"testing"
)

type fakePushRepository struct {
	sent          *models.PushNotification
	sentTopics    []string
	subscribed    []string
	unsubscribed  []string
	deviceToken   string
	err           error
	subscribeCall bool
}

func (f *fakePushRepository) SendToTopics(_ context.Context, n *models.PushNotification, topics []string) error {
	f.sent, f.sentTopics = n, topics
	return f.err
}

func (f *fakePushRepository) Subscribe(_ context.Context, token string, topics []string) error {
	f.subscribeCall, f.deviceToken, f.subscribed = true, token, topics
	return f.err
}

func (f *fakePushRepository) Unsubscribe(_ context.Context, token string, topics []string) error {
	f.deviceToken, f.unsubscribed = token, topics
	return f.err
}

func TestNotifyPostPublished(t *testing.T) {
	tests := []struct {
		name       string
		post       *requests.HookPostModelRequest
		wantTitle  string
		wantTopics []string
		wantData   map[string]string
	}{
		{
			name:       "post of a service goes to the general and the service topics",
			post:       &requests.HookPostModelRequest{ID: "p1", Title: "Nueva beca", Service: &requests.Service{ID: "svc1", ShortName: "PAE"}},
			wantTitle:  "Nueva publicación de PAE",
			wantTopics: []string{"all", "service.svc1"},
			wantData:   map[string]string{"post_id": "p1", "service_id": "svc1"},
		},
		{
			name:       "service without abbreviation keeps the generic title",
			post:       &requests.HookPostModelRequest{ID: "p1", Title: "Beca", Service: &requests.Service{ID: "svc1"}},
			wantTitle:  "Nueva publicación",
			wantTopics: []string{"all", "service.svc1"},
			wantData:   map[string]string{"post_id": "p1", "service_id": "svc1"},
		},
		{
			name:       "post without service only goes to the general topic",
			post:       &requests.HookPostModelRequest{ID: "p1", Title: "Beca"},
			wantTitle:  "Nueva publicación",
			wantTopics: []string{"all"},
			wantData:   map[string]string{"post_id": "p1"},
		},
		{
			name:       "an unsafe service id never becomes a topic",
			post:       &requests.HookPostModelRequest{ID: "p1", Title: "Beca", Service: &requests.Service{ID: "a b'] || 'x", ShortName: "PAE"}},
			wantTitle:  "Nueva publicación de PAE",
			wantTopics: []string{"all"},
			wantData:   map[string]string{"post_id": "p1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakePushRepository{}

			err := NewPushNotificationUseCase(repository).NotifyPostPublished(context.Background(), tt.post)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repository.sent.Title != tt.wantTitle || repository.sent.Body != tt.post.Title {
				t.Errorf("notification = %+v", repository.sent)
			}
			if !reflect.DeepEqual(repository.sentTopics, tt.wantTopics) || !reflect.DeepEqual(repository.sent.Data, tt.wantData) {
				t.Errorf("topics = %v, data = %v; want %v, %v", repository.sentTopics, repository.sent.Data, tt.wantTopics, tt.wantData)
			}
		})
	}
}

func TestNotifyPostPublishedPropagatesTheRepositoryError(t *testing.T) {
	errSend := errors.New("fcm down")
	repository := &fakePushRepository{err: errSend}

	err := NewPushNotificationUseCase(repository).NotifyPostPublished(context.Background(), &requests.HookPostModelRequest{ID: "p1"})

	if !errors.Is(err, errSend) {
		t.Fatalf("error = %v, want %v", err, errSend)
	}
}

func TestSubscribeValidation(t *testing.T) {
	tooMany := make([]string, 21)
	for i := range tooMany {
		tooMany[i] = "all"
	}

	tests := []struct {
		name    string
		token   string
		topics  []string
		wantErr bool
	}{
		{"general topic", "tok", []string{"all"}, false},
		{"service topics", "tok", []string{"service.abc123", "service.XYZ9"}, false},
		{"empty token", "", []string{"all"}, true},
		{"token too long", strings.Repeat("a", 4097), []string{"all"}, true},
		{"no topics", "tok", nil, true},
		{"too many topics", "tok", tooMany, true},
		{"unknown topic", "tok", []string{"news"}, true},
		{"service topic without id", "tok", []string{"service."}, true},
		{"service topic with unsafe characters", "tok", []string{"service.a/b"}, true},
		{"condition injection", "tok", []string{"all' in topics || 'x"}, true},
		{"one invalid topic invalidates the request", "tok", []string{"all", "bad"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakePushRepository{}
			useCase := NewPushNotificationUseCase(repository)

			subscribeErr := useCase.Subscribe(context.Background(), tt.token, tt.topics)
			unsubscribeErr := useCase.Unsubscribe(context.Background(), tt.token, tt.topics)

			for _, err := range []error{subscribeErr, unsubscribeErr} {
				if tt.wantErr && !errors.Is(err, ErrInvalidSubscription) {
					t.Errorf("error = %v, want ErrInvalidSubscription", err)
				}
				if !tt.wantErr && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.wantErr && (repository.subscribed != nil || repository.unsubscribed != nil) {
				t.Error("an invalid request must not reach Firebase")
			}
			if !tt.wantErr && (!reflect.DeepEqual(repository.subscribed, tt.topics) || repository.deviceToken != tt.token) {
				t.Errorf("repository received (%s, %v)", repository.deviceToken, repository.subscribed)
			}
		})
	}
}
