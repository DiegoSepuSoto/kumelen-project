package firebase

import (
	"context"
	"fmt"
	"iluvatar/src/domain/models"
	"iluvatar/src/domain/repository"

	"firebase.google.com/go/v4/auth"
)

// idTokenClient es lo que se usa de *auth.Client; permite sustituirlo en las pruebas.
type idTokenClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

type authVerifier struct {
	client idTokenClient
}

func NewAuthVerifier(client idTokenClient) *authVerifier {
	return &authVerifier{client: client}
}

// Verify valida firma, emisor, audiencia y vencimiento del ID token (lo hace el SDK de Firebase).
func (v *authVerifier) Verify(ctx context.Context, idToken string) (*models.Identity, error) {
	token, err := v.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", repository.ErrInvalidToken, err)
	}
	if token == nil || token.UID == "" {
		return nil, fmt.Errorf("%w: the token has no subject", repository.ErrInvalidToken)
	}
	email, _ := token.Claims["email"].(string)
	verified, _ := token.Claims["email_verified"].(bool)
	return &models.Identity{UID: token.UID, Email: email, EmailVerified: verified}, nil
}
