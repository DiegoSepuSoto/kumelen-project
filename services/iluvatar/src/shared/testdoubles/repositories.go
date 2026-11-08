// Package testdoubles contiene dobles de prueba escritos a mano de las interfaces de dominio.
// Cada método delega en un campo función; llamar a uno sin configurar falla con un pánico explícito.
package testdoubles

import "iluvatar/src/domain/models"

type ServiceRepository struct {
	GetServicesWithPostsFunc func() (*models.HomePage, error)
	GetDetailedServiceFunc   func(serviceID string) (*models.ServicePage, error)
	GetServicesFunc          func() ([]*models.Service, error)
	GetServiceWithPostsFunc  func(serviceID string) (*models.ServicePostsPage, error)
}

func (r *ServiceRepository) GetServicesWithPosts() (*models.HomePage, error) {
	if r.GetServicesWithPostsFunc == nil {
		panic("unexpected call to ServiceRepository.GetServicesWithPosts")
	}
	return r.GetServicesWithPostsFunc()
}

func (r *ServiceRepository) GetDetailedService(serviceID string) (*models.ServicePage, error) {
	if r.GetDetailedServiceFunc == nil {
		panic("unexpected call to ServiceRepository.GetDetailedService")
	}
	return r.GetDetailedServiceFunc(serviceID)
}

func (r *ServiceRepository) GetServices() ([]*models.Service, error) {
	if r.GetServicesFunc == nil {
		panic("unexpected call to ServiceRepository.GetServices")
	}
	return r.GetServicesFunc()
}

func (r *ServiceRepository) GetServiceWithPosts(serviceID string) (*models.ServicePostsPage, error) {
	if r.GetServiceWithPostsFunc == nil {
		panic("unexpected call to ServiceRepository.GetServiceWithPosts")
	}
	return r.GetServiceWithPostsFunc(serviceID)
}

type CategoryRepository struct {
	GetCategoriesFunc        func() ([]*models.Category, error)
	GetCategoryWithPostsFunc func(categoryID string) (*models.CategoryPostsPage, error)
}

func (r *CategoryRepository) GetCategories() ([]*models.Category, error) {
	if r.GetCategoriesFunc == nil {
		panic("unexpected call to CategoryRepository.GetCategories")
	}
	return r.GetCategoriesFunc()
}

func (r *CategoryRepository) GetCategoryWithPosts(categoryID string) (*models.CategoryPostsPage, error) {
	if r.GetCategoryWithPostsFunc == nil {
		panic("unexpected call to CategoryRepository.GetCategoryWithPosts")
	}
	return r.GetCategoryWithPostsFunc(categoryID)
}

type PostRepository struct {
	GetDetailedPostFunc func(postID string) (*models.PostPage, error)
	GetPostsByTermFunc  func(searchTerm string) (*models.PostSearchResult, error)
}

func (r *PostRepository) GetDetailedPost(postID string) (*models.PostPage, error) {
	if r.GetDetailedPostFunc == nil {
		panic("unexpected call to PostRepository.GetDetailedPost")
	}
	return r.GetDetailedPostFunc(postID)
}

func (r *PostRepository) GetPostsByTerm(searchTerm string) (*models.PostSearchResult, error) {
	if r.GetPostsByTermFunc == nil {
		panic("unexpected call to PostRepository.GetPostsByTerm")
	}
	return r.GetPostsByTermFunc(searchTerm)
}

type NotificationRepository struct {
	GetNotificationsFunc func() (*models.NotificationsPage, error)
}

func (r *NotificationRepository) GetNotifications() (*models.NotificationsPage, error) {
	if r.GetNotificationsFunc == nil {
		panic("unexpected call to NotificationRepository.GetNotifications")
	}
	return r.GetNotificationsFunc()
}
