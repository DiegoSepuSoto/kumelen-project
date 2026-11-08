package requests

// HookRequest es el webhook de Strapi 5 (https://docs.strapi.io/cms/features/webhooks).
type HookRequest struct {
	Entry interface{} `json:"entry"`
	Event string      `json:"event"`
	Model string      `json:"model"`
}

// HookPostModelRequest es la entrada (entry) de un post publicado, con el servicio poblado.
type HookPostModelRequest struct {
	ID       string   `json:"documentId"`
	Title    string   `json:"Titulo"`
	Overview string   `json:"Resumen"`
	Service  *Service `json:"servicio"`
}

type Service struct {
	ID        string `json:"documentId"`
	Name      string `json:"Nombre"`
	ShortName string `json:"Sigla"`
}
