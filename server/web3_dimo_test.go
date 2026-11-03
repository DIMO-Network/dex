package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

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
	createConfiguredWeb3Stub(t, s, id, web3StubConfig{}, grantTypes...)
}

func createConfiguredWeb3Stub(t *testing.T, s *Server, id string, stub web3StubConfig, grantTypes ...string) {
	t.Helper()
	config, err := json.Marshal(stub)
	require.NoError(t, err)
	require.NoError(t, s.storage.CreateConnector(t.Context(), storage.Connector{
		ID: id, Type: "web3Stub", Name: id, ResourceVersion: "1", Config: config, GrantTypes: grantTypes,
	}))
}

// generateChallenge starts a web3 login and returns its state.
func generateChallenge(t *testing.T, baseURL string) string {
	t.Helper()
	var gen struct {
		State string `json:"state"`
	}
	require.NoError(t, json.Unmarshal(postForm(t, baseURL+"/auth/web3/generate_challenge", challengeForm()), &gen))
	return gen.State
}

// redeemCode exchanges a public client's code for its tokens.
func redeemCode(t *testing.T, baseURL, clientID, code string) (accessToken, idToken string) {
	t.Helper()
	var tokens struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	require.NoError(t, json.Unmarshal(postForm(t, baseURL+"/token", url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {clientID},
		"redirect_uri": {testDomain},
		"code":         {code},
	}), &tokens))
	return tokens.AccessToken, tokens.IDToken
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

	status, body := postJSON(t, httpServer.URL+"/auth/web3/verify_direct", map[string]string{"signed": "0x00", "state": generateChallenge(t, httpServer.URL)})
	require.Equal(t, http.StatusForbidden, status, string(body))

	resp, err := http.PostForm(httpServer.URL+"/auth/web3/submit_challenge", url.Values{
		"client_id":  {testLicense},
		"domain":     {testDomain},
		"state":      {generateChallenge(t, httpServer.URL)},
		"grant_type": {"authorization_code"},
		"signature":  {"0x00"},
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// MFA scoped to other connector types doesn't apply to a web3 login, on this path or the
// dispatcher's.
func TestWeb3DirectLoginAllowedWhenMFADoesNotApply(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, func(c *Config) {
		c.MFAProviders = map[string]mfa.Provider{"totp": mfa.NewTOTPProvider("dex", []string{"ldap"})}
		c.DefaultMFAChain = []string{"totp"}
	})
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}, Public: true}))

	status, body := postJSON(t, httpServer.URL+"/auth/web3/verify_direct", map[string]string{"signed": "0x00", "state": generateChallenge(t, httpServer.URL)})
	require.Equal(t, http.StatusOK, status, string(body))
}

func TestWeb3DirectEndpoints(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createConfiguredWeb3Stub(t, s, "web3", web3StubConfig{Signer: testSignerLower})
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}, Public: true}))

	t.Run("verify_direct returns a code", func(t *testing.T) {
		status, body := postJSON(t, httpServer.URL+"/auth/web3/verify_direct", map[string]string{"signed": "0x00", "state": generateChallenge(t, httpServer.URL)})
		require.Equal(t, http.StatusOK, status, string(body))
		var verified struct {
			Code string `json:"code"`
		}
		require.NoError(t, json.Unmarshal(body, &verified))

		accessToken, idToken := redeemCode(t, httpServer.URL, testLicense, verified.Code)
		requireSignerClaim(t, accessToken, idToken, testSignerChecked)
	})

	t.Run("init returns a state the wallet can sign in with", func(t *testing.T) {
		var initialized struct {
			State string `json:"state"`
		}
		require.NoError(t, json.Unmarshal(postForm(t, httpServer.URL+"/auth/web3/init", url.Values{
			"client_id":     {testLicense},
			"redirect_uri":  {testDomain},
			"scope":         {"openid email"},
			"response_type": {"code"},
		}), &initialized))

		status, body := postJSON(t, httpServer.URL+"/auth/web3/challenge", map[string]string{"address": testLicense, "state": initialized.State})
		require.Equal(t, http.StatusOK, status, string(body))
	})
}

// A logged-in request holds the identity its tokens will carry, including the signer that
// token-exchange checks for revocation. Nothing may rewrite it between verify and the
// redirect the browser follows afterwards.
func TestWeb3LoginCannotChangeAfterVerify(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createConfiguredWeb3Stub(t, s, "web3", web3StubConfig{Signer: testSignerLower})
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}, Public: true}))

	state := generateChallenge(t, httpServer.URL)
	status, body := postJSON(t, httpServer.URL+"/auth/web3/verify", map[string]string{"signed": "0x00", "state": state})
	require.Equal(t, http.StatusOK, status, string(body))
	var verified struct {
		Redirect string `json:"redirect"`
	}
	require.NoError(t, json.Unmarshal(body, &verified))

	status, body = postJSON(t, httpServer.URL+"/auth/web3/challenge", map[string]string{"address": testSignerChecked, "state": state})
	require.Equal(t, http.StatusBadRequest, status, string(body))
	status, body = postJSON(t, httpServer.URL+"/auth/web3/verify", map[string]string{"signed": "0x00", "state": state})
	require.Equal(t, http.StatusBadRequest, status, string(body))

	resp, err := noRedirectClient().Get(httpServer.URL + verified.Redirect)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	location, err := resp.Location()
	require.NoError(t, err)

	accessToken, idToken := redeemCode(t, httpServer.URL, testLicense, location.Query().Get("code"))
	requireSignerClaim(t, accessToken, idToken, testSignerChecked)
}

// The second check, inside the storage update, closes the gap after the first one read.
func TestWeb3ChallengeRefusedWhenLoginCompletesFirst(t *testing.T) {
	registerWeb3Stub(t)

	var login func()
	httpServer, s := newTestServer(t, func(c *Config) {
		c.Storage = beforeAuthRequestUpdate{c.Storage, func() {
			if login != nil {
				login()
			}
		}}
	})
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	state := generateChallenge(t, httpServer.URL)
	login = func() {
		login = nil
		require.NoError(t, s.storage.UpdateAuthRequest(t.Context(), state, func(a storage.AuthRequest) (storage.AuthRequest, error) {
			a.LoggedIn = true
			return a, nil
		}))
	}

	status, body := postJSON(t, httpServer.URL+"/auth/web3/challenge", map[string]string{"address": testLicense, "state": state})
	require.Equal(t, http.StatusBadRequest, status, string(body))
}

type beforeAuthRequestUpdate struct {
	storage.Storage
	before func()
}

func (b beforeAuthRequestUpdate) UpdateAuthRequest(ctx context.Context, id string, updater func(storage.AuthRequest) (storage.AuthRequest, error)) error {
	b.before()
	return b.Storage.UpdateAuthRequest(ctx, id, updater)
}

func TestWeb3LoginRefusedOnExpiredRequest(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	state := generateChallenge(t, httpServer.URL)
	require.NoError(t, s.storage.UpdateAuthRequest(t.Context(), state, func(a storage.AuthRequest) (storage.AuthRequest, error) {
		a.Expiry = time.Now().Add(-time.Minute)
		return a, nil
	}))

	status, body := postJSON(t, httpServer.URL+"/auth/web3/challenge", map[string]string{"address": testLicense, "state": state})
	require.Equal(t, http.StatusBadRequest, status, string(body))
	status, body = postJSON(t, httpServer.URL+"/auth/web3/verify", map[string]string{"signed": "0x00", "state": state})
	require.Equal(t, http.StatusBadRequest, status, string(body))
}

func TestWeb3RejectedSignatureDoesNotLogIn(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createConfiguredWeb3Stub(t, s, "web3", web3StubConfig{Reject: true})
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	state := generateChallenge(t, httpServer.URL)
	for _, path := range []string{"/auth/web3/verify", "/auth/web3/verify_direct"} {
		status, body := postJSON(t, httpServer.URL+path, map[string]string{"signed": "0x00", "state": state})
		require.Equal(t, http.StatusBadRequest, status, string(body))
	}

	authReq, err := s.storage.GetAuthRequest(t.Context(), state)
	require.NoError(t, err)
	require.False(t, authReq.LoggedIn)
}

// The challenge endpoints run on one connector type. On any other they would overwrite the
// connector data another connector keeps on the auth request.
func TestWeb3EndpointsRefuseOtherConnectors(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	ctx := t.Context()
	require.NoError(t, s.storage.CreateClient(ctx, storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	form := challengeForm()
	form.Set("redirect_uri", testDomain)
	for _, path := range []string{"/auth/mock/generate_challenge", "/auth/mock/init"} {
		resp, err := http.PostForm(httpServer.URL+path, form)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, path)
	}

	const connectorData = `{"pkce":"verifier"}`
	authReq := storage.AuthRequest{
		ID: storage.NewID(), ClientID: testLicense, ConnectorID: "mock",
		ConnectorData: []byte(connectorData), Expiry: time.Now().Add(time.Minute),
	}
	require.NoError(t, s.storage.CreateAuthRequest(ctx, authReq))

	for _, path := range []string{"/auth/mock/challenge", "/auth/web3/challenge", "/auth/does-not-exist/challenge"} {
		status, body := postJSON(t, httpServer.URL+path, map[string]string{"address": testLicense, "state": authReq.ID})
		require.Equal(t, http.StatusBadRequest, status, path+": "+string(body))
	}

	stored, err := s.storage.GetAuthRequest(ctx, authReq.ID)
	require.NoError(t, err)
	require.JSONEq(t, connectorData, string(stored.ConnectorData))
}

// An auth request belongs to the connector it started on, even if another one could verify
// the wallet.
func TestWeb3RequestStaysOnItsConnector(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	createWeb3Stub(t, s, "web3-other")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	state := generateChallenge(t, httpServer.URL)

	status, body := postJSON(t, httpServer.URL+"/auth/web3-other/challenge", map[string]string{"address": testLicense, "state": state})
	require.Equal(t, http.StatusBadRequest, status, string(body))
	status, body = postJSON(t, httpServer.URL+"/auth/web3-other/verify", map[string]string{"signed": "0x00", "state": state})
	require.Equal(t, http.StatusBadRequest, status, string(body))
}

type failingAuthRequestUpdates struct{ storage.Storage }

func (failingAuthRequestUpdates) UpdateAuthRequest(context.Context, string, func(storage.AuthRequest) (storage.AuthRequest, error)) error {
	return errors.New("database down")
}

func TestWeb3ChallengeReportsStorageFailure(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, func(c *Config) {
		c.Storage = failingAuthRequestUpdates{c.Storage}
	})
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	status, body := postJSON(t, httpServer.URL+"/auth/web3/challenge", map[string]string{"address": testLicense, "state": generateChallenge(t, httpServer.URL)})
	require.Equal(t, http.StatusInternalServerError, status, string(body))
}

func TestWeb3ChallengeResponsesAreJSON(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()
	createWeb3Stub(t, s, "web3")
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}}))

	resp, err := http.PostForm(httpServer.URL+"/auth/web3/generate_challenge", challengeForm())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	payload, err := json.Marshal(map[string]string{"address": testLicense, "state": generateChallenge(t, httpServer.URL)})
	require.NoError(t, err)
	resp, err = http.Post(httpServer.URL+"/auth/web3/challenge", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

// A signature proves a signer is still a signer, and only a login that made one can carry
// the signer in its token, so a session can't stand in for it.
func TestWeb3LoginNeverResumesSession(t *testing.T) {
	registerWeb3Stub(t)

	httpServer, s := newTestServerWithSessions(t, nil)
	defer httpServer.Close()
	createConfiguredWeb3Stub(t, s, "web3", web3StubConfig{Signer: testSignerLower})
	require.NoError(t, s.storage.CreateClient(t.Context(), storage.Client{ID: testLicense, RedirectURIs: []string{testDomain}, Public: true}))

	payload, err := json.Marshal(map[string]string{"signed": "0x00", "state": generateChallenge(t, httpServer.URL)})
	require.NoError(t, err)
	verifyResp, err := http.Post(httpServer.URL+"/auth/web3/verify", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer verifyResp.Body.Close()
	require.Equal(t, http.StatusOK, verifyResp.StatusCode)
	require.NotEmpty(t, verifyResp.Cookies(), "verify didn't start a session")

	login, err := http.NewRequest(http.MethodGet, httpServer.URL+"/auth/web3?"+url.Values{
		"client_id":     {testLicense},
		"redirect_uri":  {testDomain},
		"scope":         {"openid email"},
		"response_type": {"code"},
	}.Encode(), nil)
	require.NoError(t, err)
	for _, cookie := range verifyResp.Cookies() {
		login.AddCookie(cookie)
	}

	resp, err := noRedirectClient().Do(login)
	require.NoError(t, err)
	defer resp.Body.Close()
	page, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(page))
	require.Contains(t, string(page), `id="tmpl-auth-id"`)
}

func TestDIMOConnectorsRegistered(t *testing.T) {
	require.Contains(t, ConnectorsConfig, "web3")
	require.Contains(t, ConnectorsConfig, "apple")
}

// Wallets and the developer SDKs call these endpoints from other origins.
func TestWeb3EndpointsAnswerCORS(t *testing.T) {
	const origin = "https://app.example"

	httpServer, _ := newTestServer(t, func(c *Config) {
		c.AllowedOrigins = []string{origin}
	})
	defer httpServer.Close()

	for _, endpoint := range []string{"generate_challenge", "submit_challenge", "challenge", "verify", "verify_direct", "init"} {
		t.Run(endpoint, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodOptions, httpServer.URL+"/auth/web3/"+endpoint, nil)
			require.NoError(t, err)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, origin, resp.Header.Get("Access-Control-Allow-Origin"))
		})
	}
}

// The lifetime override has to reach every path that signs tokens: grants hand the issuer
// the client they authenticated, and implicit responses have to load it.
func TestClientLifetimeOverrideReachesTokens(t *testing.T) {
	const (
		clientID = "short-lived"
		override = 30 * time.Minute
	)

	httpServer, _ := newTestServer(t, func(c *Config) {
		c.SupportedResponseTypes = []string{"code", "id_token"}
		c.Storage = storage.WithStaticClients(c.Storage, []storage.Client{
			{ID: clientID, RedirectURIs: []string{testDomain}, Public: true, IDTokenExpiry: override.String()},
		})
	})
	defer httpServer.Close()

	// authorize runs the login and returns where it sends the browser back to.
	authorize := func(t *testing.T, responseType string) *url.URL {
		t.Helper()
		var done *url.URL
		client := &http.Client{CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if req.URL.Host == "example.com" {
				done = req.URL
				return http.ErrUseLastResponse
			}
			return nil
		}}
		resp, err := client.Get(httpServer.URL + "/auth/mock?" + url.Values{
			"client_id":     {clientID},
			"redirect_uri":  {testDomain},
			"scope":         {"openid"},
			"response_type": {responseType},
			"nonce":         {"nonce"},
		}.Encode())
		require.NoError(t, err)
		defer resp.Body.Close()
		require.NotNil(t, done, "login never returned to the client")
		return done
	}

	requireLifetime := func(t *testing.T, idToken string) {
		t.Helper()
		claims := tokenClaims(t, idToken)
		require.Equal(t, override.Seconds(), claims["exp"].(float64)-claims["iat"].(float64))
	}

	t.Run("code", func(t *testing.T) {
		_, idToken := redeemCode(t, httpServer.URL, clientID, authorize(t, "code").Query().Get("code"))
		requireLifetime(t, idToken)
	})
	t.Run("implicit", func(t *testing.T) {
		fragment, err := url.ParseQuery(authorize(t, "id_token").Fragment)
		require.NoError(t, err)
		requireLifetime(t, fragment.Get("id_token"))
	})
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
