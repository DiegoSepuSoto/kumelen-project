package post

import (
	"errors"
	"iluvatar/src/domain/models"
	"iluvatar/src/shared/testdoubles"
	"testing"
)

func TestGetPostPage(t *testing.T) {
	page := &models.PostPage{}
	errRepo := errors.New("post not found")

	tests := []struct {
		name    string
		repoErr error
		want    *models.PostPage
	}{
		{name: "returns the post page", want: page},
		{name: "propagates the repository error", repoErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID string
			useCase := NewPostUseCase(&testdoubles.PostRepository{
				GetDetailedPostFunc: func(postID string) (*models.PostPage, error) {
					gotID = postID
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return page, nil
				},
			})

			got, err := useCase.GetPostPage("42")

			if gotID != "42" {
				t.Errorf("repository received id %q, want %q", gotID, "42")
			}
			if !errors.Is(err, tt.repoErr) || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.repoErr)
			}
		})
	}
}

func TestGetPostsByTerm(t *testing.T) {
	result := &models.PostSearchResult{Posts: []*models.PostForPreview{{ID: "1", Title: "Beca"}}}
	errRepo := errors.New("search failed")

	tests := []struct {
		name    string
		repoErr error
		want    *models.PostSearchResult
	}{
		{name: "returns the search result", want: result},
		{name: "propagates the repository error", repoErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotTerm string
			useCase := NewPostUseCase(&testdoubles.PostRepository{
				GetPostsByTermFunc: func(searchTerm string) (*models.PostSearchResult, error) {
					gotTerm = searchTerm
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return result, nil
				},
			})

			got, err := useCase.GetPostsByTerm("beca")

			if gotTerm != "beca" {
				t.Errorf("repository received term %q, want %q", gotTerm, "beca")
			}
			if !errors.Is(err, tt.repoErr) || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.repoErr)
			}
		})
	}
}
