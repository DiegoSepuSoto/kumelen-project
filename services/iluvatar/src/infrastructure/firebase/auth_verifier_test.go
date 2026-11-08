package firebase

import (
	"context"
	"errors"
	"iluvatar/src/domain/repository"
	"testing"

	"firebase.google.com/go/v4/auth"
)

type fakeIDTokenClient struct {
	token *auth.Token
	err   error
	got   string
}

func (f *fakeIDTokenClient) VerifyIDToken(_ context.Context, idToken string) (*auth.Token, error) {
	f.got = idToken
	return f.token, f.err
}

func TestVerify(t *testing.T) {
	t.Run("returns the uid of a valid token", func(t *testing.T) {
		client := &fakeIDTokenClient{token: &auth.Token{UID: "uid-1"}}

		identity, err := NewAuthVerifier(client).Verify(context.Background(), "the-token")

		if err != nil || identity.UID != "uid-1" || client.got != "the-token" {
			t.Fatalf("got (%+v, %v), token sent to firebase = %q", identity, err, client.got)
		}
	})

	failures := map[string]*fakeIDTokenClient{
		"firebase rejects the token": {err: errors.New("ID token has expired")},
		"token without subject":      {token: &auth.Token{}},
		"nil token without error":    {},
	}
	for name, client := range failures {
		t.Run(name, func(t *testing.T) {
			identity, err := NewAuthVerifier(client).Verify(context.Background(), "x")

			if !errors.Is(err, repository.ErrInvalidToken) || identity != nil {
				t.Errorf("got (%+v, %v), want ErrInvalidToken", identity, err)
			}
		})
	}
}
