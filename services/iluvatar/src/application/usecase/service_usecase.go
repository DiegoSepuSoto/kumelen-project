package usecase

import "iluvatar/src/domain/models"

type ServiceUseCase interface {
	GetServicePage(serviceID string) (*models.ServicePage, error)
	GetServicePostsPage(serviceID string) (*models.ServicePostsPage, error)
}
