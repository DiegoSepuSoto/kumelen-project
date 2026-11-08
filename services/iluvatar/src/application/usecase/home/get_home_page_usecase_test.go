package home

import (
	"errors"
	"iluvatar/src/domain/models"
	"iluvatar/src/shared/testdoubles"
	"testing"
)

func TestGetHomePage(t *testing.T) {
	page := &models.HomePage{ServicesWithPosts: []*models.ServiceWithPosts{{ID: "1", Abbreviation: "PAE"}}}
	errRepo := errors.New("cms down")

	t.Run("returns the page from the repository", func(t *testing.T) {
		useCase := NewHomeUseCase(&testdoubles.ServiceRepository{
			GetServicesWithPostsFunc: func() (*models.HomePage, error) { return page, nil },
		})

		got, err := useCase.GetHomePage()

		if err != nil || got != page {
			t.Fatalf("got (%v, %v), want (%v, nil)", got, err, page)
		}
	})

	t.Run("propagates the repository error", func(t *testing.T) {
		useCase := NewHomeUseCase(&testdoubles.ServiceRepository{
			GetServicesWithPostsFunc: func() (*models.HomePage, error) { return nil, errRepo },
		})

		got, err := useCase.GetHomePage()

		if !errors.Is(err, errRepo) || got != nil {
			t.Fatalf("got (%v, %v), want (nil, %v)", got, err, errRepo)
		}
	})
}
