package post

import "iluvatar/src/domain/repository"

type UseCase struct {
	postRepository repository.PostRepository
}

func NewPostUseCase(postRepository repository.PostRepository) *UseCase {
	return &UseCase{postRepository: postRepository}
}
