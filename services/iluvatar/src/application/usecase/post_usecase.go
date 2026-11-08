package usecase

import "iluvatar/src/domain/models"

type PostUseCase interface {
	GetPostPage(postID string) (*models.PostPage, error)
	GetPostsByTerm(searchTerm string) (*models.PostSearchResult, error)
}
