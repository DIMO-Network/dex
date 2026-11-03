package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/connector/apple/secret"
)

const staleSecret = "stale"

// payloadKeySet accepts any signature, so tests can write ID tokens by hand.
type payloadKeySet struct{}

func (payloadKeySet) VerifySignature(_ context.Context, jwt string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.Split(jwt, ".")[1])
}

func idToken(t *testing.T, claims map[string]any) string {
	t.Helper()
	claims["iss"] = "https://appleid.apple.com"
	claims["sub"] = "apple-user"
	claims["email"] = "user@example.com"
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	encode := base64.RawURLEncoding.EncodeToString
	return encode([]byte(`{"alg":"RS256"}`)) + "." + encode(payload) + "." + encode([]byte("signature"))
}

// expiredSecret returns a client secret handler whose secret is always due for rotation.
func expiredSecret(t *testing.T) *secret.Secret {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	keyFile := filepath.Join(t.TempDir(), "key.pem")
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o600))

	s, err := secret.NewSecret(&secret.Config{
		TeamID: "team", KeyID: "key", PrivateKeyFile: keyFile, ClientID: "client",
		SecretDuration: 60, SecretExpiryMin: 120,
	})
	require.NoError(t, err)
	return s
}

// newTestConnector returns a connector against a token endpoint that issues claims, and the
// count of requests that endpoint received.
func newTestConnector(t *testing.T, claims map[string]any, skipEmailVerified bool) (*oidcConnector, *atomic.Int32) {
	t.Helper()
	tokenRequests := new(atomic.Int32)
	rawIDToken := idToken(t, claims)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRequests.Add(1)
		if got := r.PostFormValue("client_secret"); got == staleSecret {
			http.Error(w, `{"error":"invalid_client"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]string{
			"access_token": "access", "token_type": "Bearer", "refresh_token": "refresh", "id_token": rawIDToken,
		}))
	}))
	t.Cleanup(server.Close)

	return &oidcConnector{
		redirectURI: "https://dex.example/callback",
		oauth2Config: &oauth2.Config{
			ClientID:     "client",
			ClientSecret: staleSecret,
			Endpoint:     oauth2.Endpoint{TokenURL: server.URL, AuthStyle: oauth2.AuthStyleInParams},
			Scopes:       []string{oidc.ScopeOpenID, "email"},
		},
		verifier: oidc.NewVerifier("", payloadKeySet{}, &oidc.Config{
			SkipClientIDCheck: true, SkipExpiryCheck: true, SkipIssuerCheck: true,
		}),
		logger:                    slog.New(slog.DiscardHandler),
		secret:                    expiredSecret(t),
		insecureSkipEmailVerified: skipEmailVerified,
	}, tokenRequests
}

func callback(form url.Values) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

func TestEmailVerifiedClaim(t *testing.T) {
	tests := []struct {
		name         string
		claim        any
		skipVerified bool
		want         bool
		wantErr      bool
	}{
		{name: "string true", claim: "true", want: true},
		{name: "string false", claim: "false", want: false},
		{name: "boolean true", claim: true, want: true},
		{name: "boolean false", claim: false, want: false},
		{name: "boolean false isn't overridden", claim: false, skipVerified: true, want: false},
		{name: "missing", wantErr: true},
		{name: "missing, overridden", skipVerified: true, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims := map[string]any{}
			if tc.claim != nil {
				claims["email_verified"] = tc.claim
			}
			conn, _ := newTestConnector(t, claims, tc.skipVerified)

			identity, err := conn.HandlePOST(connector.Scopes{}, callback(url.Values{"code": {"code"}}))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, identity.EmailVerified)
		})
	}
}

func TestHandlePOSTStopsOnProviderError(t *testing.T) {
	conn, tokenRequests := newTestConnector(t, map[string]any{"email_verified": "true"}, false)

	_, err := conn.HandlePOST(connector.Scopes{}, callback(url.Values{"error": {"access_denied"}}))
	require.ErrorContains(t, err, "access_denied")
	require.Zero(t, tokenRequests.Load())
}

// The login that rotates the secret may land on another replica than Apple's callback, and
// a refresh may come long after any login, so each path has to rotate it for itself.
func TestTokenRequestsRotateExpiredSecret(t *testing.T) {
	claims := map[string]any{"email_verified": "true"}

	t.Run("callback", func(t *testing.T) {
		conn, _ := newTestConnector(t, claims, false)
		_, err := conn.HandlePOST(connector.Scopes{}, callback(url.Values{"code": {"code"}}))
		require.NoError(t, err)
	})
	t.Run("refresh", func(t *testing.T) {
		conn, _ := newTestConnector(t, claims, false)
		data, err := json.Marshal(connectorData{RefreshToken: []byte("refresh")})
		require.NoError(t, err)
		_, err = conn.Refresh(t.Context(), connector.Scopes{}, connector.Identity{ConnectorData: data})
		require.NoError(t, err)
	})
}

func TestConcurrentLoginsShareSecret(t *testing.T) {
	conn, _ := newTestConnector(t, map[string]any{"email_verified": "true"}, false)

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			_, err := conn.LoginURL(connector.Scopes{}, conn.redirectURI, "state")
			assert.NoError(t, err)
			_, err = conn.HandlePOST(connector.Scopes{}, callback(url.Values{"code": {"code"}}))
			assert.NoError(t, err)
		})
	}
	wg.Wait()
}
