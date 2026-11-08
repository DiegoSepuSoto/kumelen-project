// Package firebase conecta Laurelin con Firebase Auth: verifica los ID tokens de quienes administran el registro.
package firebase

import (
	"context"
	"errors"
	"fmt"

	firebaseSDK "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// NewApp crea la app de Firebase. Verificar tokens solo necesita el ID del proyecto.
func NewApp(ctx context.Context, projectID string, opts ...option.ClientOption) (*firebaseSDK.App, error) {
	if projectID == "" {
		return nil, errors.New("FIREBASE_PROJECT_ID is required")
	}
	app, err := firebaseSDK.NewApp(ctx, &firebaseSDK.Config{ProjectID: projectID}, opts...)
	if err != nil {
		return nil, fmt.Errorf("can't initialize the firebase app: %w", err)
	}
	return app, nil
}
