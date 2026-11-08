package strapi

import (
	"iluvatar/src/domain/models"
	"net/url"
)

// notificationsLimit es la cantidad de avisos recientes que muestra la pantalla de notificaciones.
const notificationsLimit = "10"

type notificationRepository struct {
	client *Client
}

func NewNotificationRepository(client *Client) *notificationRepository {
	return &notificationRepository{client: client}
}

// GetNotifications devuelve los avisos más recientes; en el CMS los avisos son el contenido tipo "Aviso".
func (r *notificationRepository) GetNotifications() (*models.NotificationsPage, error) {
	query := url.Values{}
	addFields(query, "", "Titulo", "Cuerpo")
	addFields(query, "populate[servicio]", "Nombre")
	addFields(query, "populate[servicio][populate][Logo]", "url")
	addSort(query, "", "createdAt:desc")
	query.Set("pagination[limit]", notificationsLimit)

	var notices listResponse[avisoEntity]
	if err := r.client.get("avisos", query, &notices); err != nil {
		return nil, err
	}

	notifications := make([]*models.Notification, 0, len(notices.Data))
	for _, notice := range notices.Data {
		notification := &models.Notification{
			ID:    notice.DocumentID,
			Title: notice.Titulo,
			Body:  notice.Cuerpo,
		}
		if notice.Servicio != nil {
			notification.ServiceName = notice.Servicio.Nombre
			notification.ServiceImageURL = r.client.mediaURL(notice.Servicio.Logo)
		}
		notifications = append(notifications, notification)
	}
	return &models.NotificationsPage{Notifications: notifications}, nil
}
