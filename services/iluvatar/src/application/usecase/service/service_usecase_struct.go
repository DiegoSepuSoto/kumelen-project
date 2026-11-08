package service

import "iluvatar/src/domain/repository"

type UseCase struct {
	serviceRepository repository.ServiceRepository
}

func NewServiceUseCase(serviceRepository repository.ServiceRepository) *UseCase {
	return &UseCase{serviceRepository: serviceRepository}
}
