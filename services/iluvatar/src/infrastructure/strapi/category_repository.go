package strapi

import (
	"iluvatar/src/domain/models"
	"net/url"
)

type categoryRepository struct {
	client *Client
}

func NewCategoryRepository(client *Client) *categoryRepository {
	return &categoryRepository{client: client}
}

func (r *categoryRepository) GetCategories() ([]*models.Category, error) {
	query := url.Values{}
	addFields(query, "", "Titulo")
	addSort(query, "", "Titulo:asc")
	query.Set("pagination[limit]", maxPageSize)

	var categories listResponse[categoriaEntity]
	if err := r.client.get("categorias", query, &categories); err != nil {
		return nil, err
	}

	result := make([]*models.Category, 0, len(categories.Data))
	for _, category := range categories.Data {
		result = append(result, &models.Category{ID: category.DocumentID, Name: category.Titulo})
	}
	return result, nil
}

func (r *categoryRepository) GetCategoryWithPosts(categoryID string) (*models.CategoryPostsPage, error) {
	query := url.Values{}
	addFields(query, "", "Titulo")
	addPostPreview(query, "populate[posts]")
	addSort(query, "populate[posts]", "createdAt:desc")

	var response itemResponse[categoriaEntity]
	if err := r.client.get("categorias/"+url.PathEscape(categoryID), query, &response); err != nil {
		return nil, err
	}
	if response.Data == nil {
		return nil, ErrNotFound
	}

	return &models.CategoryPostsPage{
		ID:              response.Data.DocumentID,
		Name:            response.Data.Titulo,
		PostsForPreview: r.client.postsForPreview(response.Data.Posts),
	}, nil
}
