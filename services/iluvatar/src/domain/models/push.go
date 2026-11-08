package models

const (
	// TopicAll recibe todas las publicaciones.
	TopicAll = "all"

	serviceTopicPrefix = "service."
)

// ServiceTopic es el topic de los dispositivos que siguen a un servicio estudiantil (su documentId en el CMS).
func ServiceTopic(serviceID string) string {
	return serviceTopicPrefix + serviceID
}

// PushNotification es el mensaje que se envía a los dispositivos suscritos.
type PushNotification struct {
	Title string
	Body  string
	Data  map[string]string
}
