// Package laurelin consulta al servicio Laurelin (registro de estudiantes) para saber si un correo es de un
// estudiante activo. Guarda las respuestas en memoria un tiempo y, si Laurelin cae, sigue aceptando a quien ya
// había sido verificado como estudiante.
package laurelin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"iluvatar/src/domain/models"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	lookupPath      = "/internal/v1/students/lookup"
	maxResponseSize = 1 << 20
	lookupTimeout   = 5 * time.Second
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
	now     func() time.Time

	studentTTL    time.Duration // cuánto vale una respuesta "es estudiante"
	notStudentTTL time.Duration // cuánto vale "no es estudiante" (corto: que un alta nueva se note pronto)
	staleFor      time.Duration // hasta cuándo se acepta una respuesta "es estudiante" vencida si Laurelin cae
	maxEntries    int

	mu      sync.Mutex
	entries map[string]entry
	group   singleflight.Group
}

type entry struct {
	standing  models.StudentStanding
	fetchedAt time.Time
}

type Option func(*Client)

// WithCache ajusta los tiempos de la caché. Con todos en cero no se guarda nada.
func WithCache(studentTTL, notStudentTTL, staleFor time.Duration, maxEntries int) Option {
	return func(c *Client) {
		c.studentTTL, c.notStudentTTL, c.staleFor, c.maxEntries = studentTTL, notStudentTTL, staleFor, maxEntries
	}
}

func NewClient(baseURL, token string, httpClient *http.Client, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    httpClient,
		now:     time.Now,
		entries: map[string]entry{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// key es el hash del correo: en memoria no queda ningún correo en claro.
func key(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(sum[:])
}

func (c *Client) ttl(standing models.StudentStanding) time.Duration {
	if standing.IsStudent {
		return c.studentTTL
	}
	return c.notStudentTTL
}

func (c *Client) cached(k string) (entry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[k]
	return e, ok
}

// Lookup implementa repository.StudentDirectory.
func (c *Client) Lookup(ctx context.Context, email string) (*models.StudentStanding, error) {
	k := key(email)
	previous, hadPrevious := c.cached(k)
	if hadPrevious && c.now().Sub(previous.fetchedAt) < c.ttl(previous.standing) {
		standing := previous.standing
		return &standing, nil
	}

	// Una sola consulta por correo a la vez; sigue aunque el cliente que la pidió se desconecte.
	result, err, _ := c.group.Do(k, func() (any, error) {
		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), lookupTimeout)
		defer cancel()
		standing, err := c.fetch(fetchCtx, email)
		if err != nil {
			return nil, err
		}
		c.store(k, *standing)
		return standing, nil
	})
	if err != nil {
		// Laurelin no responde: quien ya era estudiante sigue entrando por un tiempo. Quien no, no.
		if hadPrevious && previous.standing.IsStudent && c.now().Sub(previous.fetchedAt) < c.staleFor {
			slog.Warn("laurelin unavailable, using the last known answer", "error", err.Error())
			standing := previous.standing
			return &standing, nil
		}
		return nil, err
	}
	standing := *result.(*models.StudentStanding)
	return &standing, nil
}

func (c *Client) store(k string, standing models.StudentStanding) {
	if c.ttl(standing) <= 0 && c.staleFor <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxEntries > 0 && len(c.entries) >= c.maxEntries {
		now := c.now()
		for existing, e := range c.entries {
			if now.Sub(e.fetchedAt) >= c.staleFor {
				delete(c.entries, existing)
			}
		}
		for existing := range c.entries { // sigue lleno: se descarta una entrada cualquiera
			if len(c.entries) < c.maxEntries {
				break
			}
			delete(c.entries, existing)
		}
	}
	c.entries[k] = entry{standing: standing, fetchedAt: c.now()}
}

type lookupResponse struct {
	IsStudent       bool   `json:"is_student"`
	Status          string `json:"status"`
	Level           string `json:"level"`
	InstitutionName string `json:"institution_name"`
	ProgramName     string `json:"program_name"`
}

// fetch consulta a Laurelin. El correo va en el cuerpo, no en la URL, y los errores no lo incluyen.
func (c *Client) fetch(ctx context.Context, email string) (*models.StudentStanding, error) {
	body, err := json.Marshal(map[string]string{"email": email})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+lookupPath, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("can't build the laurelin request")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.token)

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("laurelin request failed: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusNotFound:
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseSize))
		return &models.StudentStanding{}, nil
	case http.StatusOK:
		var parsed lookupResponse
		if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize)).Decode(&parsed); err != nil {
			return nil, errors.New("laurelin sent an unreadable response")
		}
		return &models.StudentStanding{
			IsStudent: parsed.IsStudent, Status: parsed.Status, Level: parsed.Level,
			InstitutionName: parsed.InstitutionName, ProgramName: parsed.ProgramName,
		}, nil
	default:
		// 400 (correo con formato inválido), 401 (token de servicio incorrecto), 503 (sin token configurado) o 5xx.
		return nil, fmt.Errorf("laurelin responded with status %d", response.StatusCode)
	}
}
