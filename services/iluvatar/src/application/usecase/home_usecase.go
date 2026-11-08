package usecase

import "iluvatar/src/domain/models"

type HomeUseCase interface {
	GetHomePage() (*models.HomePage, error)
}
