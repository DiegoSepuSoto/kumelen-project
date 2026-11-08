package usecase

import "iluvatar/src/domain/models"

type CatalogUseCase interface {
	GetCatalogPage() (*models.CatalogPage, error)
}
