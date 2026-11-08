package category

import (
	"errors"
	"iluvatar/src/domain/models"
	"iluvatar/src/shared/testdoubles"
	"testing"
)

func TestGetCategoryPostsPage(t *testing.T) {
	page := &models.CategoryPostsPage{}
	errRepo := errors.New("category not found")

	tests := []struct {
		name    string
		repoErr error
		want    *models.CategoryPostsPage
	}{
		{name: "returns the category posts page", want: page},
		{name: "propagates the repository error", repoErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID string
			useCase := NewCategoryUseCase(&testdoubles.CategoryRepository{
				GetCategoryWithPostsFunc: func(categoryID string) (*models.CategoryPostsPage, error) {
					gotID = categoryID
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return page, nil
				},
			})

			got, err := useCase.GetCategoryPostsPage("3")

			if gotID != "3" {
				t.Errorf("repository received id %q, want %q", gotID, "3")
			}
			if !errors.Is(err, tt.repoErr) || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.repoErr)
			}
		})
	}
}
