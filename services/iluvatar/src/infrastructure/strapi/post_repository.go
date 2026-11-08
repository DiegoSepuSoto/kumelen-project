package strapi

import (
	"iluvatar/src/domain/models"
	"net/url"
)

type postRepository struct {
	client *Client
}

func NewPostRepository(client *Client) *postRepository {
	return &postRepository{client: client}
}

func (r *postRepository) GetDetailedPost(postID string) (*models.PostPage, error) {
	query := url.Values{}
	addFields(query, "", "Titulo", "Cuerpo")
	addFields(query, "populate[Imagen]", "url")
	addFields(query, "populate[categorias]", "Titulo")
	addFields(query, "populate[servicio]", "Nombre")

	var response itemResponse[postEntity]
	if err := r.client.get("posts/"+url.PathEscape(postID), query, &response); err != nil {
		return nil, err
	}
	if response.Data == nil {
		return nil, ErrNotFound
	}
	post := response.Data

	categories := make([]*models.PostCategory, 0, len(post.Categorias))
	for _, category := range post.Categorias {
		categories = append(categories, &models.PostCategory{ID: category.DocumentID, Name: category.Titulo})
	}

	page := &models.PostPage{
		ID:         post.DocumentID,
		Title:      post.Titulo,
		Image:      r.client.mediaURL(post.Imagen),
		Body:       post.Cuerpo,
		Categories: categories,
	}
	if post.Servicio != nil {
		page.ServiceID = post.Servicio.DocumentID
		page.ServiceName = post.Servicio.Nombre
	}
	return page, nil
}

// GetPostsByTerm busca el término en el cuerpo de los posts, sin distinguir mayúsculas.
func (r *postRepository) GetPostsByTerm(searchTerm string) (*models.PostSearchResult, error) {
	query := url.Values{}
	query.Set("filters[Cuerpo][$containsi]", searchTerm)
	addPostPreview(query, "")
	addSort(query, "", "createdAt:desc")
	query.Set("pagination[limit]", maxPageSize)

	var posts listResponse[postEntity]
	if err := r.client.get("posts", query, &posts); err != nil {
		return nil, err
	}
	return &models.PostSearchResult{Posts: r.client.postsForPreview(posts.Data)}, nil
}

// postsForPreview convierte posts del CMS en tarjetas; siempre devuelve una lista (nunca nil).
func (c *Client) postsForPreview(posts []postEntity) []*models.PostForPreview {
	previews := make([]*models.PostForPreview, 0, len(posts))
	for _, post := range posts {
		previews = append(previews, &models.PostForPreview{
			ID:       post.DocumentID,
			Title:    post.Titulo,
			Summary:  post.Resumen,
			ImageURL: c.mediaURL(post.Imagen),
		})
	}
	return previews
}
