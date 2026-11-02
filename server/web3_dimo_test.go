package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/server/connectors"
	"github.com/dexidp/dex/server/mfa"
	"github.com/dexidp/dex/storage"
)

func noRedirectClient() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func postJSON(t *testing.T, endpoint string, body any) (int, []byte) {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

func createWeb3Stub(t *testing.T, s *Server, id string, grantTypes ...string) {
	t.Helper()
	config, err := json.Marshal(web3StubConfig{})
	require.NoError(t, err)
	require.NoError(t, s.storage.CreateConnector(t.Context(), storage.Connector{
		ID: id, Type: "web3Stub", Name: id, ResourceVersion: "1", Config: config, GrantTypes: grantTypes,
	}))
}

func challengeForm() url.Values {
	return url.Values{
		"client_id":     {testLicense},
		"domain":        {testDomain},
		"scope":         {"openid email"},
		"response_type": {"code"},
		"address":       {testLicense},
	}
}

func TestWeb3BrowserLogin(t *testing.T) {
	t.Run("without sessions", func(t *testing.T) {
		httpServer, s := newTestServer(t, nil)
		defer httpServer.Close()
		testWeb3BrowserLogin(t, httpServer, s, false)
	})
	t.Run("with sessions", func(t *testing.T) {
		httpServer, s := newTestServerWithSessions(t, nil)
		defer httpServer.Close()
		testWeb3BrowserLogin(t, httpServer, s, true)
	})
}

func testWeb3BrowserLogin(t *testing.T, httpServer *httptest.Server, s *Server, sessions bool) {
	registerWeb3Stub(t)
	createWeb3Stub(t, s, "web3")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}, Public: true}))

	login := url.Values{
		"client_id":     {testLicense},
		"redirect_uri":  {testDomain},
		"scope":         {"openid email"},
		"response_type": {"code"},
		"state":         {"client-state"},
	}
	resp, err := http.Get(httpServer.URL + "/auth/web3?" + login.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()
	page, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(page))
	require.Contains(t, string(page), `id="tmpl-challenge-url" value="/auth/web3/challenge"`)
	require.Contains(t, string(page), `id="tmpl-verify-url" value="/auth/web3/verify"`)

	match := regexp.MustCompile(`id="tmpl-auth-id" value="([^"]+)"`).FindSubmatch(page)
	require.NotNil(t, match, "web3 page has no auth request id")
	state := string(match[1])

	status, body := postJSON(t, httpServer.URL+"/auth/web3/challenge", map[string]string{"address": testLicense, "state": state})
	require.Equal(t, http.StatusOK, status, string(body))

	payload, err := json.Marshal(map[string]string{"signed": "0x00", "state": state})
	require.NoError(t, err)
	verifyResp, err := http.Post(httpServer.URL+"/auth/web3/verify", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer verifyResp.Body.Close()
	require.Equal(t, http.StatusOK, verifyResp.StatusCode)
	var verified struct {
		Redirect string `json:"redirect"`
	}
	require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&verified))
	require.True(t, strings.HasPrefix(verified.Redirect, "/auth?"), verified.Redirect)

	// The redirect is the page's next navigation; the dispatcher issues the code from it.
	dispatch, err := http.NewRequest(http.MethodGet, httpServer.URL+verified.Redirect, nil)
	require.NoError(t, err)
	if sessions {
		require.NotEmpty(t, verifyResp.Cookies(), "verify didn't start a session")
		for _, cookie := range verifyResp.Cookies() {
			dispatch.AddCookie(cookie)
		}
	}
	resp, err = noRedirectClient().Do(dispatch)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	location, err := resp.Location()
	require.NoError(t, err)
	require.Equal(t, "client-state", location.Query().Get("state"))

	var tokens struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	require.NoError(t, json.Unmarshal(postForm(t, httpServer.URL+"/token", url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {testLicense},
		"redirect_uri": {testDomain},
		"code":         {location.Query().Get("code")},
	}), &tokens))

	claims := tokenClaims(t, tokens.IDToken)
	require.Equal(t, "web3", claims["provider_id"])
	require.Equal(t, testLicense, claims["ethereum_address"])
	if sessions {
		require.NotEmpty(t, claims["sid"], "a browser login names its session")
	} else {
		require.NotContains(t, claims, "sid")
	}
}

// The web3 endpoints bind an auth request to the connector in their path, so they owe the
// same checks handleConnectorLogin makes: otherwise a client restricted to other connectors,
// or a connector closed to the grant, could still be logged in through them.
func TestWeb3EndpointsEnforceConnectorPolicy(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	createWeb3Stub(t, s, "web3-no-code", "refresh_token")

	const restricted = "0x71efD5d71a597eB6BEC28DFDB05a49283a3e20c5"
	ctx := t.Context()
	require.NoError(t, s.storage.CreateClient(ctx, storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))
	require.NoError(t, s.storage.CreateClient(ctx, storage.Client{ID: restricted, RedirectURIs: []string{testDomain}, AllowedConnectors: []string{"mock"}}))

	withClient := func(id string) url.Values {
		form := challengeForm()
		form.Set("client_id", id)
		form.Set("redirect_uri", testDomain)
		return form
	}

	tests := []struct {
		name string
		path string
		form url.Values
		want int
	}{
		{name: "generate_challenge, client not allowed", path: "/auth/web3/generate_challenge", form: withClient(restricted), want: http.StatusForbidden},
		{name: "init, client not allowed", path: "/auth/web3/init", form: withClient(restricted), want: http.StatusForbidden},
		{name: "generate_challenge, grant not allowed", path: "/auth/web3-no-code/generate_challenge", form: withClient(testLicense), want: http.StatusBadRequest},
		{name: "init, grant not allowed", path: "/auth/web3-no-code/init", form: withClient(testLicense), want: http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.PostForm(httpServer.URL+tc.path, tc.form)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tc.want, resp.StatusCode)
		})
	}

	resp, err := http.PostForm(httpServer.URL+"/auth/web3/generate_challenge", withClient(testLicense))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWeb3ChallengeRejectsMalformedAddress(t *testing.T) {
	httpServer, _ := newTestServer(t, nil)
	defer httpServer.Close()

	status, body := postJSON(t, httpServer.URL+"/auth/mock/challenge", map[string]string{"address": "not-an-address", "state": "x"})
	require.Equal(t, http.StatusBadRequest, status, string(body))
}

// The direct endpoints never return to the /auth dispatcher, which is where MFA runs.
func TestWeb3DirectLoginRefusedWhenClientRequiresMFA(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, func(c *Config) {
		c.MFAProviders = map[string]mfa.Provider{"totp": mfa.NewTOTPProvider("dex", nil)}
		c.DefaultMFAChain = []string{"totp"}
	})
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	generate := func() string {
		var gen struct {
			State string `json:"state"`
		}
		require.NoError(t, json.Unmarshal(postForm(t, httpServer.URL+"/auth/web3/generate_challenge", challengeForm()), &gen))
		return gen.State
	}

	status, body := postJSON(t, httpServer.URL+"/auth/web3/verify_direct", map[string]string{"signed": "0x00", "state": generate()})
	require.Equal(t, http.StatusForbidden, status, string(body))

	resp, err := http.PostForm(httpServer.URL+"/auth/web3/submit_challenge", url.Values{
		"client_id":  {testLicense},
		"domain":     {testDomain},
		"state":      {generate()},
		"grant_type": {"authorization_code"},
		"signature":  {"0x00"},
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

type formPOSTStubConfig struct{}

func (*formPOSTStubConfig) Open(string, *slog.Logger) (connector.Connector, error) {
	return formPOSTStub{}, nil
}

type formPOSTStub struct{}

func (formPOSTStub) LoginURL(_ connector.Scopes, callbackURL, state string) (string, error) {
	return "https://provider.example/auth?" + url.Values{"redirect_uri": {callbackURL}, "state": {state}}.Encode(), nil
}

func (formPOSTStub) HandlePOST(_ connector.Scopes, r *http.Request) (connector.Identity, error) {
	return connector.Identity{UserID: r.PostFormValue("user"), Username: "apple-user"}, nil
}

func TestFormPOSTConnectorLogin(t *testing.T) {
	ConnectorsConfig["formPOSTStub"] = func() connectors.ConnectorConfig { return new(formPOSTStubConfig) }
	t.Cleanup(func() { delete(ConnectorsConfig, "formPOSTStub") })

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	ctx := context.Background()
	require.NoError(t, s.storage.CreateConnector(ctx, storage.Connector{ID: "apple", Type: "formPOSTStub", Name: "Apple", ResourceVersion: "1"}))
	require.NoError(t, s.storage.CreateClient(ctx, storage.Client{ID: "app", RedirectURIs: []string{testDomain}, Public: true}))

	client := noRedirectClient()
	resp, err := client.Get(httpServer.URL + "/auth/apple?" + url.Values{
		"client_id":     {"app"},
		"redirect_uri":  {testDomain},
		"scope":         {"openid"},
		"response_type": {"code"},
	}.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusFound, resp.StatusCode)
	providerURL, err := resp.Location()
	require.NoError(t, err)
	require.Equal(t, "provider.example", providerURL.Host)
	state := providerURL.Query().Get("state")

	// The connector only answers POSTs, so a GET with a valid state is refused.
	resp, err = client.Get(httpServer.URL + "/callback?state=" + state)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = client.PostForm(httpServer.URL+"/callback", url.Values{"state": {state}, "user": {"apple-id"}})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	continueURL, err := resp.Location()
	require.NoError(t, err)

	resp, err = client.Get(httpServer.URL + continueURL.RequestURI())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	done, err := resp.Location()
	require.NoError(t, err)
	require.NotEmpty(t, done.Query().Get("code"), done.String())
}
