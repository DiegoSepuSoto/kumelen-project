package strapi

import (
	"iluvatar/src/domain/models"
	"net/url"
	"sync"
)

// homePostsPerService es la cantidad de posts recientes que muestra cada servicio en la pantalla de inicio.
const homePostsPerService = "5"

type serviceRepository struct {
	client *Client
}

func NewServiceRepository(client *Client) *serviceRepository {
	return &serviceRepository{client: client}
}

// GetServicesWithPosts arma la pantalla de inicio: los servicios con sus posts más recientes.
// Strapi 5 no permite limitar la cantidad de posts dentro de populate, por lo que se consulta
// cada servicio en paralelo.
func (r *serviceRepository) GetServicesWithPosts() (*models.HomePage, error) {
	query := url.Values{}
	addFields(query, "", "Sigla")
	addSort(query, "", "createdAt:asc")
	query.Set("pagination[limit]", maxPageSize)

	var services listResponse[servicioEntity]
	if err := r.client.get("servicios", query, &services); err != nil {
		return nil, err
	}

	posts := make([][]postEntity, len(services.Data))
	errs := make([]error, len(services.Data))
	var wg sync.WaitGroup
	for i, service := range services.Data {
		wg.Add(1)
		go func() {
			defer wg.Done()
			posts[i], errs[i] = r.latestPosts(service.DocumentID)
		}()
	}
	wg.Wait()

	servicesWithPosts := make([]*models.ServiceWithPosts, 0, len(services.Data))
	for i, service := range services.Data {
		if errs[i] != nil {
			return nil, errs[i]
		}
		if len(posts[i]) == 0 {
			continue
		}
		servicesWithPosts = append(servicesWithPosts, &models.ServiceWithPosts{
			ID:              service.DocumentID,
			Abbreviation:    service.Sigla,
			PostsForPreview: r.client.postsForPreview(posts[i]),
		})
	}

	return &models.HomePage{ServicesWithPosts: servicesWithPosts}, nil
}

func (r *serviceRepository) latestPosts(serviceID string) ([]postEntity, error) {
	query := url.Values{}
	query.Set("filters[servicio][documentId][$eq]", serviceID)
	addPostPreview(query, "")
	addSort(query, "", "createdAt:desc")
	query.Set("pagination[limit]", homePostsPerService)

	var posts listResponse[postEntity]
	if err := r.client.get("posts", query, &posts); err != nil {
		return nil, err
	}
	return posts.Data, nil
}

func (r *serviceRepository) GetDetailedService(serviceID string) (*models.ServicePage, error) {
	query := url.Values{}
	addFields(query, "", "Nombre", "Descripcion")
	addFields(query, "populate[Logo]", "url")
	addFields(query, "populate[personas]", "Nombre", "Cargo", "Correo")
	addFields(query, "populate[personas][populate][Imagen]", "url")
	addFields(query, "populate[contactos]", "Nombre", "Tipo", "Link")

	var response itemResponse[servicioEntity]
	if err := r.client.get("servicios/"+url.PathEscape(serviceID), query, &response); err != nil {
		return nil, err
	}
	if response.Data == nil {
		return nil, ErrNotFound
	}
	service := response.Data

	persons := make([]*models.Person, 0, len(service.Personas))
	for _, person := range service.Personas {
		persons = append(persons, &models.Person{
			ID:             person.DocumentID,
			Name:           person.Nombre,
			Charge:         person.Cargo,
			Email:          person.Correo,
			ProfilePicture: r.client.mediaURL(person.Imagen),
		})
	}

	contacts := make([]*models.Contact, 0, len(service.Contactos))
	for _, contact := range service.Contactos {
		contacts = append(contacts, &models.Contact{
			ID:   contact.DocumentID,
			Name: contact.Nombre,
			Type: contact.Tipo,
			Link: contact.Link,
		})
	}

	return &models.ServicePage{
		ID:          service.DocumentID,
		LogoURL:     r.client.mediaURL(service.Logo),
		Name:        service.Nombre,
		Description: service.Descripcion,
		Persons:     persons,
		Contacts:    contacts,
	}, nil
}

func (r *serviceRepository) GetServices() ([]*models.Service, error) {
	query := url.Values{}
	addFields(query, "", "Nombre")
	addFields(query, "populate[Logo]", "url")
	addSort(query, "", "Nombre:asc")
	query.Set("pagination[limit]", maxPageSize)

	var services listResponse[servicioEntity]
	if err := r.client.get("servicios", query, &services); err != nil {
		return nil, err
	}

	result := make([]*models.Service, 0, len(services.Data))
	for _, service := range services.Data {
		result = append(result, &models.Service{
			ID:      service.DocumentID,
			Name:    service.Nombre,
			LogoURL: r.client.mediaURL(service.Logo),
		})
	}
	return result, nil
}

func (r *serviceRepository) GetServiceWithPosts(serviceID string) (*models.ServicePostsPage, error) {
	query := url.Values{}
	addFields(query, "", "Nombre")
	addPostPreview(query, "populate[posts]")
	addSort(query, "populate[posts]", "createdAt:desc")

	var response itemResponse[servicioEntity]
	if err := r.client.get("servicios/"+url.PathEscape(serviceID), query, &response); err != nil {
		return nil, err
	}
	if response.Data == nil {
		return nil, ErrNotFound
	}

	return &models.ServicePostsPage{
		ID:              response.Data.DocumentID,
		Name:            response.Data.Nombre,
		PostsForPreview: r.client.postsForPreview(response.Data.Posts),
	}, nil
}
