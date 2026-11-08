package service

import (
	"errors"
	"iluvatar/src/domain/models"
	"iluvatar/src/shared/testdoubles"
	"testing"
)

func TestGetServicePage(t *testing.T) {
	page := &models.ServicePage{}
	errRepo := errors.New("service not found")

	tests := []struct {
		name    string
		repoErr error
		want    *models.ServicePage
	}{
		{name: "returns the service page", want: page},
		{name: "propagates the repository error", repoErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID string
			useCase := NewServiceUseCase(&testdoubles.ServiceRepository{
				GetDetailedServiceFunc: func(serviceID string) (*models.ServicePage, error) {
					gotID = serviceID
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return page, nil
				},
			})

			got, err := useCase.GetServicePage("7")

			if gotID != "7" {
				t.Errorf("repository received id %q, want %q", gotID, "7")
			}
			if !errors.Is(err, tt.repoErr) || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.repoErr)
			}
		})
	}
}

func TestGetServicePostsPage(t *testing.T) {
	page := &models.ServicePostsPage{}
	errRepo := errors.New("posts failed")

	tests := []struct {
		name    string
		repoErr error
		want    *models.ServicePostsPage
	}{
		{name: "returns the service posts page", want: page},
		{name: "propagates the repository error", repoErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID string
			useCase := NewServiceUseCase(&testdoubles.ServiceRepository{
				GetServiceWithPostsFunc: func(serviceID string) (*models.ServicePostsPage, error) {
					gotID = serviceID
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return page, nil
				},
			})

			got, err := useCase.GetServicePostsPage("7")

			if gotID != "7" {
				t.Errorf("repository received id %q, want %q", gotID, "7")
			}
			if !errors.Is(err, tt.repoErr) || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.repoErr)
			}
		})
	}
}
