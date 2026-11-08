package post

import "iluvatar/src/domain/models"

func (u *UseCase) GetPostPage(postID string) (*models.PostPage, error) {
	return u.postRepository.GetDetailedPost(postID)
}
