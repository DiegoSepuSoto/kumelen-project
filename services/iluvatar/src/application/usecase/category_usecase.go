package usecase

import "iluvatar/src/domain/models"

type CategoryUseCase interface {
	GetCategoryPostsPage(categoryID string) (*models.CategoryPostsPage, error)
}
