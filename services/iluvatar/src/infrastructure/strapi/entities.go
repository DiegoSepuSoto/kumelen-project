package strapi

// Tipos de la respuesta REST de Strapi 5: los campos van planos y cada documento se identifica por documentId.

type media struct {
	URL string `json:"url"`
}

type servicioEntity struct {
	DocumentID  string           `json:"documentId"`
	Nombre      string           `json:"Nombre"`
	Sigla       string           `json:"Sigla"`
	Descripcion string           `json:"Descripcion"`
	Logo        *media           `json:"Logo"`
	Posts       []postEntity     `json:"posts"`
	Personas    []personaEntity  `json:"personas"`
	Contactos   []contactoEntity `json:"contactos"`
}

type postEntity struct {
	DocumentID string            `json:"documentId"`
	Titulo     string            `json:"Titulo"`
	Resumen    string            `json:"Resumen"`
	Cuerpo     string            `json:"Cuerpo"`
	Imagen     *media            `json:"Imagen"`
	Categorias []categoriaEntity `json:"categorias"`
	Servicio   *servicioEntity   `json:"servicio"`
}

type categoriaEntity struct {
	DocumentID string       `json:"documentId"`
	Titulo     string       `json:"Titulo"`
	Posts      []postEntity `json:"posts"`
}

type personaEntity struct {
	DocumentID string `json:"documentId"`
	Nombre     string `json:"Nombre"`
	Cargo      string `json:"Cargo"`
	Correo     string `json:"Correo"`
	Imagen     *media `json:"Imagen"`
}

type contactoEntity struct {
	DocumentID string `json:"documentId"`
	Nombre     string `json:"Nombre"`
	Tipo       string `json:"Tipo"`
	Link       string `json:"Link"`
}

type avisoEntity struct {
	DocumentID string          `json:"documentId"`
	Titulo     string          `json:"Titulo"`
	Cuerpo     string          `json:"Cuerpo"`
	Servicio   *servicioEntity `json:"servicio"`
}

type listResponse[T any] struct {
	Data []T `json:"data"`
}

// itemResponse usa un puntero para distinguir "data": null.
type itemResponse[T any] struct {
	Data *T `json:"data"`
}
