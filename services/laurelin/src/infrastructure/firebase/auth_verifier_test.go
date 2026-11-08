package firebase

import (
	"context"
	"errors"
	"testing"

	"laurelin/src/domain/repository"

	"firebase.google.com/go/v4/auth"
)

type fakeClient struct {
	token *auth.Token
	err   error
}

func (f fakeClient) VerifyIDToken(context.Context, string) (*auth.Token, error) {
	return f.token, f.err
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name         string
		client       fakeClient
		wantErr      bool
		wantEmail    string
		wantVerified bool
	}{
		{"verified email", fakeClient{token: &auth.Token{UID: "u1", Claims: map[string]any{"email": "a@utem.cl", "email_verified": true}}}, false, "a@utem.cl", true},
		{"unverified email", fakeClient{token: &auth.Token{UID: "u1", Claims: map[string]any{"email": "a@utem.cl", "email_verified": false}}}, false, "a@utem.cl", false},
		{"no email claim", fakeClient{token: &auth.Token{UID: "u1", Claims: map[string]any{}}}, false, "", false},
		{"claims of the wrong type", fakeClient{token: &auth.Token{UID: "u1", Claims: map[string]any{"email": 5, "email_verified": "yes"}}}, false, "", false},
		{"sdk rejects the token", fakeClient{err: errors.New("expired")}, true, "", false},
		{"token without subject", fakeClient{token: &auth.Token{}}, true, "", false},
		{"nil token", fakeClient{}, true, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, err := NewAuthVerifier(tt.client).Verify(context.Background(), "token")
			if tt.wantErr {
				if !errors.Is(err, repository.ErrInvalidToken) {
					t.Fatalf("err = %v, want ErrInvalidToken", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if identity.Email != tt.wantEmail || identity.EmailVerified != tt.wantVerified {
				t.Errorf("identity = %+v", identity)
			}
		})
	}
}
