// Package strapi implementa los repositorios del dominio sobre la API REST de Strapi 5.
package strapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"encoding/json"

	"golang.org/x/sync/singleflight"
)

// ErrNotFound indica que el CMS respondió 404 para el recurso pedido.
var ErrNotFound = errors.New("resource not found in cms")

const maxResponseBytes = 10 << 20

type Client struct {
	baseURL string
	// publicURL es la dirección del CMS que ve el navegador (para las imágenes). Si el CMS se consulta por una
	// red interna (http://cms:1337) no coincide con baseURL; vacía significa que es la misma.
	publicURL  string
	token      string
	httpClient *http.Client

	// Caché de respuestas: sirve las consultas repetidas sin molestar al CMS y, si el CMS falla,
	// entrega la última respuesta buena (aunque esté vencida) durante staleFor. El contenido es el mismo
	// para todas las personas autenticadas, por eso se comparte.
	cacheTTL   time.Duration
	staleFor   time.Duration
	maxEntries int
	now        func() time.Time
	mu         sync.Mutex
	cache      map[string]cacheEntry
	group      singleflight.Group
}

type cacheEntry struct {
	body     []byte
	storedAt time.Time
}

type Option func(*Client)

// WithCache activa la caché: una respuesta se considera fresca durante ttl y, si el CMS falla,
// se reutiliza hasta staleFor después de guardada. maxEntries acota la memoria.
func WithCache(ttl, staleFor time.Duration, maxEntries int) Option {
	return func(c *Client) {
		c.cacheTTL, c.staleFor, c.maxEntries = ttl, staleFor, maxEntries
	}
}

// WithPublicURL fija la dirección pública del CMS, usada solo para construir las URLs de imágenes.
func WithPublicURL(publicURL string) Option {
	return func(c *Client) { c.publicURL = strings.TrimRight(publicURL, "/") }
}

// NewClient crea un cliente para el CMS en baseURL (sin /api). token es un API token de solo lectura.
func NewClient(baseURL, token string, httpClient *http.Client, options ...Option) *Client {
	client := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: httpClient,
		now:        time.Now,
		cache:      map[string]cacheEntry{},
	}
	for _, option := range options {
		option(client)
	}
	return client
}

// get consulta GET {baseURL}/api/{path}?{query} y decodifica el JSON de la respuesta en out.
func (c *Client) get(path string, query url.Values, out interface{}) error {
	endpoint := c.baseURL + "/api/" + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	body, err := c.fetchWithCache(path, endpoint)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("can't decode cms response for %s: %w", path, err)
	}
	return nil
}

func (c *Client) fetchWithCache(path, endpoint string) ([]byte, error) {
	if c.cacheTTL <= 0 {
		return c.fetch(path, endpoint)
	}

	if entry, ok := c.lookup(endpoint); ok && c.now().Sub(entry.storedAt) <= c.cacheTTL {
		return entry.body, nil
	}

	// Las peticiones simultáneas por lo mismo comparten una sola consulta al CMS.
	result, err, _ := c.group.Do(endpoint, func() (interface{}, error) {
		body, err := c.fetch(path, endpoint)
		if err == nil {
			c.store(endpoint, body)
		}
		return body, err
	})
	if err == nil {
		return result.([]byte), nil
	}

	if errors.Is(err, ErrNotFound) {
		c.remove(endpoint)
		return nil, err
	}
	if entry, ok := c.lookup(endpoint); ok && c.now().Sub(entry.storedAt) <= c.staleFor {
		slog.Warn("serving a stale cms response", "path", path, "age_seconds", int(c.now().Sub(entry.storedAt).Seconds()), "error", err.Error())
		return entry.body, nil
	}
	return nil, err
}

func (c *Client) fetch(path, endpoint string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	if c.token != "" {
		request.Header.Set("Authorization", "Bearer "+c.token)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	switch {
	case response.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case response.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("cms returned status %d for %s", response.StatusCode, path)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxResponseBytes {
		return nil, fmt.Errorf("cms response for %s is too large", path)
	}
	return body, nil
}

func (c *Client) lookup(key string) (cacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.cache[key]
	return entry, ok
}

func (c *Client) remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

func (c *Client) store(key string, body []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.cache[key]; !exists && c.maxEntries > 0 && len(c.cache) >= c.maxEntries {
		c.evictLocked()
	}
	c.cache[key] = cacheEntry{body: body, storedAt: c.now()}
}

// evictLocked libera espacio: primero lo que ya no sirve ni como respaldo y, si no hay, lo más antiguo.
func (c *Client) evictLocked() {
	var oldestKey string
	var oldest time.Time
	for key, entry := range c.cache {
		if c.now().Sub(entry.storedAt) > c.staleFor {
			delete(c.cache, key)
			return
		}
		if oldestKey == "" || entry.storedAt.Before(oldest) {
			oldestKey, oldest = key, entry.storedAt
		}
	}
	delete(c.cache, oldestKey)
}

// Ping comprueba que el CMS responde (Strapi publica /_health); lo usa la verificación de disponibilidad.
func (c *Client) Ping(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/_health", nil)
	if err != nil {
		return err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("cms health check returned status %d", response.StatusCode)
	}
	return nil
}

// mediaURL devuelve la URL absoluta de un archivo del CMS. Los proveedores de almacenamiento remoto
// entregan una URL completa; el almacenamiento local, una ruta relativa al CMS.
func (c *Client) mediaURL(m *media) string {
	if m == nil || m.URL == "" {
		return ""
	}
	if strings.HasPrefix(m.URL, "http://") || strings.HasPrefix(m.URL, "https://") {
		return m.URL
	}
	if c.publicURL != "" {
		return c.publicURL + m.URL
	}
	return c.baseURL + m.URL
}
