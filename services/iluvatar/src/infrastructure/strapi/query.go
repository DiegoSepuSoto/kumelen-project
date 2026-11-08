package strapi

import (
	"net/url"
	"strconv"
)

// maxPageSize es el máximo que Strapi entrega por consulta (pagination.maxLimit por defecto).
const maxPageSize = "100"

// addFields restringe los campos devueltos: path "" para el recurso raíz o "populate[posts]" para una relación.
func addFields(query url.Values, path string, names ...string) {
	for i, name := range names {
		key := "fields[" + strconv.Itoa(i) + "]"
		if path != "" {
			key = path + "[fields][" + strconv.Itoa(i) + "]"
		}
		query.Set(key, name)
	}
}

// addSort ordena por un campo; path "" para el recurso raíz o "populate[posts]" para una relación.
func addSort(query url.Values, path, order string) {
	key := "sort[0]"
	if path != "" {
		key = path + "[sort][0]"
	}
	query.Set(key, order)
}

// addPostPreview pide lo necesario para una tarjeta de post: título, resumen e imagen.
// path "" para consultar /posts o "populate[posts]" para posts anidados.
func addPostPreview(query url.Values, path string) {
	addFields(query, path, "Titulo", "Resumen")
	imagePath := "populate[Imagen]"
	if path != "" {
		imagePath = path + "[populate][Imagen]"
	}
	addFields(query, imagePath, "url")
}
