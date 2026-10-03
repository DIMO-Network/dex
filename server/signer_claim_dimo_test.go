package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/storage"
)

const (
	testLicense       = "0x299671D2b32ED62Cc61ce65D8f2b9e4f78486B37"
	testSignerLower   = "0x71efd5d71a597eb6bec28dfdb05a49283a3e20c5"
	testSignerChecked = "0x71efD5d71a597eB6BEC28DFDB05a49283a3e20c5"
	testDomain        = "https://example.com/callback"
)

func tokenClaims(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))
	return claims
}

func requireSignerClaim(t *testing.T, accessToken, idToken, want string) {
	t.Helper()
	for name, token := range map[string]string{"access token": accessToken, "ID token": idToken} {
		claims := tokenClaims(t, token)
		require.Equal(t, testLicense, claims["ethereum_address"], name)
		if want == "" {
			require.NotContains(t, claims, "signer_address", name)
		} else {
			require.Equal(t, want, claims["signer_address"], name)
		}
	}
}

func TestExchangeAuthCodeSignerAddressClaim(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, s := newTestServer(ctx, t, nil)
	defer httpServer.Close()

	client := storage.Client{ID: testLicense, Secret: "secret", RedirectURIs: []string{testDomain}}
	require.NoError(t, s.storage.CreateClient(ctx, client))

	tests := []struct {
		name          string
		connectorData []byte
		want          string
	}{
		{name: "web3 signer data adds the checksummed claim", connectorData: []byte(`{"signer_address":"` + testSignerLower + `"}`), want: testSignerChecked},
		{name: "no connector data omits the claim"},
		{name: "another connector's data omits the claim", connectorData: []byte(`{"RefreshToken":"abc"}`)},
		{name: "a malformed address omits the claim", connectorData: []byte(`{"signer_address":"0x1234"}`)},
		{name: "non-JSON connector data omits the claim", connectorData: []byte("opaque")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authCode := storage.AuthCode{
				ID:            storage.NewID(),
				ClientID:      client.ID,
				RedirectURI:   testDomain,
				ConnectorID:   "mock",
				Scopes:        []string{"openid", "email"},
				Claims:        storage.Claims{UserID: client.ID, Username: client.ID},
				Expiry:        time.Now().Add(time.Minute),
				ConnectorData: tc.connectorData,
			}
			require.NoError(t, s.storage.CreateAuthCode(ctx, authCode))

			resp, err := s.exchangeAuthCode(ctx, httptest.NewRecorder(), authCode, client)
			require.NoError(t, err)

			requireSignerClaim(t, resp.AccessToken, resp.IDToken, tc.want)
		})
	}
}

// web3StubConfig opens a Web3 connector that accepts any signature and reports Signer as
// the challenge signer, so the generate_challenge → submit_challenge → token chain runs
// end to end without a blockchain.
type web3StubConfig struct {
	Signer string `json:"signer"`
}

func (c *web3StubConfig) Open(string, *slog.Logger) (connector.Connector, error) {
	return web3Stub{signer: c.Signer}, nil
}

type web3Stub struct{ signer string }

func (web3Stub) InfuraID() string { return "" }

func (w web3Stub) Verify(address, _, _ string) (connector.Identity, error) {
	identity := connector.Identity{UserID: address, Username: address}
	if w.signer != "" {
		data, err := json.Marshal(connector.Web3SignerData{SignerAddress: w.signer})
		if err != nil {
			return identity, err
		}
		identity.ConnectorData = data
	}
	return identity, nil
}

func postForm(t *testing.T, endpoint string, form url.Values) []byte {
	t.Helper()
	resp, err := http.PostForm(endpoint, form)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	return body
}

func TestSubmitChallengeSignerAddressClaim(t *testing.T) {
	ConnectorsConfig["web3Stub"] = func() ConnectorConfig { return new(web3StubConfig) }
	t.Cleanup(func() { delete(ConnectorsConfig, "web3Stub") })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, s := newTestServer(ctx, t, nil)
	defer httpServer.Close()

	// Developer-license clients have no secret.
	require.NoError(t, s.storage.CreateClient(ctx, storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	tests := []struct {
		name        string
		connectorID string
		signer      string
		want        string
	}{
		{name: "an ERC-1271 developer-license login carries the signer", connectorID: "web3-signer", signer: testSignerLower, want: testSignerChecked},
		{name: "a login without a recovered signer carries no claim", connectorID: "web3-plain"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := json.Marshal(web3StubConfig{Signer: tc.signer})
			require.NoError(t, err)
			require.NoError(t, s.storage.CreateConnector(ctx, storage.Connector{
				ID: tc.connectorID, Type: "web3Stub", Name: tc.connectorID, ResourceVersion: "1", Config: config,
			}))

			var gen struct {
				State     string `json:"state"`
				Challenge string `json:"challenge"`
			}
			require.NoError(t, json.Unmarshal(postForm(t, httpServer.URL+"/auth/"+tc.connectorID+"/generate_challenge", url.Values{
				"client_id":     {testLicense},
				"domain":        {testDomain},
				"scope":         {"openid email"},
				"response_type": {"code"},
				"address":       {testLicense},
			}), &gen))
			require.NotEmpty(t, gen.State)

			var tokens struct {
				AccessToken string `json:"access_token"`
				IDToken     string `json:"id_token"`
			}
			require.NoError(t, json.Unmarshal(postForm(t, httpServer.URL+"/auth/"+tc.connectorID+"/submit_challenge", url.Values{
				"client_id":  {testLicense},
				"domain":     {testDomain},
				"state":      {gen.State},
				"grant_type": {"authorization_code"},
				"signature":  {"0x00"},
			}), &tokens))

			requireSignerClaim(t, tokens.AccessToken, tokens.IDToken, tc.want)
		})
	}
}
