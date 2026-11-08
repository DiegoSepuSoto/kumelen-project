//go:build firebase_emulator

package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iluvatar/src/domain/repository"
	"net/http"
	"os"
	"testing"
)

// Prueba el verificador real (SDK de Firebase) contra el emulador de Firebase Auth:
//
//	firebase emulators:start --only auth --project demo-kumelen
//	FIREBASE_AUTH_EMULATOR_HOST=127.0.0.1:9099 go test -tags firebase_emulator -run TestEmulator ./src/infrastructure/firebase/
//
// El emulador no firma los tokens, así que esta prueba cubre el cableado, la audiencia (proyecto),
// el formato y los errores, pero no la verificación de firma, que solo ocurre contra Firebase real.
const emulatorProject = "demo-kumelen"

func signUp(t *testing.T, host string) (idToken, uid string) {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"email": fmt.Sprintf("estudiante-%d@example.com", os.Getpid()), "password": "clave-de-prueba-123", "returnSecureToken": true,
	})
	response, err := http.Post("http://"+host+"/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var result struct {
		IDToken string `json:"idToken"`
		LocalID string `json:"localId"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil || result.IDToken == "" {
		t.Fatalf("sign up failed: status %d, err %v", response.StatusCode, err)
	}
	return result.IDToken, result.LocalID
}

func newVerifier(t *testing.T, projectID string) *authVerifier {
	t.Helper()
	app, err := NewApp(context.Background(), projectID)
	if err != nil {
		t.Fatal(err)
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return NewAuthVerifier(client)
}

func TestEmulatorVerifiesRealFirebaseTokens(t *testing.T) {
	host := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if host == "" {
		t.Skip("FIREBASE_AUTH_EMULATOR_HOST no está definida")
	}
	idToken, uid := signUp(t, host)

	t.Run("a token issued by Firebase for the project is accepted", func(t *testing.T) {
		identity, err := newVerifier(t, emulatorProject).Verify(context.Background(), idToken)
		if err != nil || identity.UID != uid {
			t.Fatalf("got (%+v, %v), want uid %s", identity, err, uid)
		}
	})

	t.Run("a token for another project is rejected", func(t *testing.T) {
		identity, err := newVerifier(t, "otro-proyecto").Verify(context.Background(), idToken)
		if !errors.Is(err, repository.ErrInvalidToken) || identity != nil {
			t.Fatalf("got (%+v, %v), want ErrInvalidToken", identity, err)
		}
	})

	for name, token := range map[string]string{"garbage": "not-a-token", "empty": "", "truncated": idToken[:len(idToken)/2]} {
		t.Run(name+" is rejected", func(t *testing.T) {
			if _, err := newVerifier(t, emulatorProject).Verify(context.Background(), token); !errors.Is(err, repository.ErrInvalidToken) {
				t.Fatalf("error = %v, want ErrInvalidToken", err)
			}
		})
	}
}
