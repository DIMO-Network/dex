package tokens

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dexidp/dex/server/signer"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/memory"
)

const testIssuer = "https://auth.example.com"

func newDIMOTestIssuer(t *testing.T, s storage.Storage, validFor time.Duration) *Issuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	sig, err := signer.NewMockSigner(key)
	require.NoError(t, err)
	issuerURL, err := url.Parse(testIssuer)
	require.NoError(t, err)
	return NewIssuer(s, sig, *issuerURL, validFor, time.Now, slog.New(slog.DiscardHandler))
}

func decodeClaims(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))
	return claims
}

type failingClients struct{ storage.Storage }

func (failingClients) GetClient(context.Context, string) (storage.Client, error) {
	return storage.Client{}, errors.New("database down")
}

func TestClientTokenLifetime(t *testing.T) {
	const global = 2 * time.Hour

	tests := []struct {
		name    string
		expiry  string
		absent  bool
		wantFor time.Duration
	}{
		{name: "no override uses the global lifetime", wantFor: global},
		{name: "a shorter override wins", expiry: "1h", wantFor: time.Hour},
		{name: "an override can't extend the global lifetime", expiry: "3h", wantFor: global},
		{name: "an invalid override falls back to global", expiry: "soon", wantFor: global},
		{name: "a client missing from storage uses the global lifetime", absent: true, wantFor: global},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := memory.New(slog.New(slog.DiscardHandler))
			if !tc.absent {
				store = storage.WithStaticClients(store, []storage.Client{{ID: "client", IDTokenExpiry: tc.expiry}})
			}
			issuer := newDIMOTestIssuer(t, store, global)

			token, expiry, err := issuer.SignIDToken(t.Context(), Authorization{Client: storage.Client{ID: "client"}, Scopes: []string{ScopeOpenID}}, "", "")
			require.NoError(t, err)

			claims := decodeClaims(t, token)
			require.Equal(t, tc.wantFor.Seconds(), claims["exp"].(float64)-claims["iat"].(float64))
			require.Equal(t, int64(claims["exp"].(float64)), expiry.Unix())
		})
	}
}

// A lifetime that can't be read must not silently become the longer global one.
func TestSignIDTokenFailsWhenClientUnreadable(t *testing.T) {
	store := failingClients{memory.New(slog.New(slog.DiscardHandler))}
	issuer := newDIMOTestIssuer(t, store, time.Hour)

	_, _, err := issuer.SignIDToken(t.Context(), Authorization{Client: storage.Client{ID: "client"}}, "", "")
	require.Error(t, err)
}

func TestDIMOClaims(t *testing.T) {
	const address = "0x299671D2b32ED62Cc61ce65D8f2b9e4f78486B37"

	tests := []struct {
		name         string
		userID       string
		scopes       []string
		wantEthereum any
	}{
		{name: "an address user with the email scope gets the claim", userID: address, scopes: []string{ScopeOpenID, ScopeEmail}, wantEthereum: address},
		{name: "without the email scope it is omitted", userID: address, scopes: []string{ScopeOpenID}},
		{name: "a non-address user has none", userID: "1234", scopes: []string{ScopeOpenID, ScopeEmail}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issuer := newDIMOTestIssuer(t, memory.New(slog.New(slog.DiscardHandler)), time.Hour)

			token, _, err := issuer.SignIDToken(t.Context(), Authorization{
				Client:      storage.Client{ID: "client"},
				Claims:      storage.Claims{UserID: tc.userID},
				Scopes:      tc.scopes,
				ConnectorID: "web3",
			}, "", "")
			require.NoError(t, err)

			claims := decodeClaims(t, token)
			require.Equal(t, "web3", claims["provider_id"])
			require.Equal(t, tc.wantEthereum, claims["ethereum_address"])
		})
	}
}

func TestSignCustomToken(t *testing.T) {
	issuer := newDIMOTestIssuer(t, memory.New(slog.New(slog.DiscardHandler)), time.Hour)

	token, err := issuer.SignCustomToken(t.Context(), "0xabc", []string{"dimo.zone"}, map[string]any{
		"privilege_ids": []any{1.0, 4.0},
		"sub":           "spoofed",
		"iss":           "spoofed",
	})
	require.NoError(t, err)

	claims := decodeClaims(t, token)

	// Base claims win over anything passed in as a custom claim.
	require.Equal(t, "0xabc", claims["sub"])
	require.Equal(t, testIssuer, claims["iss"])
	require.Equal(t, []any{"dimo.zone"}, claims["aud"])
	require.Equal(t, []any{1.0, 4.0}, claims["privilege_ids"])
	require.Equal(t, time.Hour.Seconds(), claims["exp"].(float64)-claims["iat"].(float64))
}
