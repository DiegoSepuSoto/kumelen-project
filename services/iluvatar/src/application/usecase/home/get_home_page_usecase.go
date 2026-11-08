package home

import "iluvatar/src/domain/models"

func (u *UseCase) GetHomePage() (*models.HomePage, error) {
	return u.serviceRepository.GetServicesWithPosts()
}
