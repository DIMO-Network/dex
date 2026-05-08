package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/storage"
)

func TestHandleInvalidFormPOSTCallbacks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, server := newTestServer(ctx, t, func(c *Config) {
		c.Storage = &emptyStorage{c.Storage}
	})
	defer httpServer.Close()

	type formPOST struct {
		State string
	}
	tests := []struct {
		FormPOST     formPOST
		ExpectedCode int
	}{
		{formPOST{}, http.StatusBadRequest},
		{formPOST{State: "abcd"}, http.StatusBadRequest},
	}

	rr := httptest.NewRecorder()

	for i, r := range tests {
		jsonValue, err := json.Marshal(r.FormPOST)
		if err != nil {
			t.Fatal(err.Error())
		}
		server.ServeHTTP(rr, httptest.NewRequest("POST", "/callback", bytes.NewBuffer(jsonValue)))
		if rr.Code != r.ExpectedCode {
			t.Fatalf("test %d expected %d, got %d", i, r.ExpectedCode, rr.Code)
		}
	}
}

type stubWeb3Connector struct{}

func (stubWeb3Connector) InfuraID() string { return "" }

func (stubWeb3Connector) Verify(address, msg, signedMsg string) (connector.Identity, error) {
	return connector.Identity{
		UserID:        "test-user-id",
		Username:      "test-user",
		Email:         "test@example.com",
		EmailVerified: true,
	}, nil
}

type updateAuthRequestErrStorage struct {
	storage.Storage
}

func (s *updateAuthRequestErrStorage) UpdateAuthRequest(id string, updater func(storage.AuthRequest) (storage.AuthRequest, error)) error {
	return errors.New("simulated update auth request error")
}

// TestHandleSubmitChallengeReturnsAfterFinalizeLoginError verifies that when
// finalizeLogin returns an error, the response contains only the error JSON
// and the handler does not fall through to the success path (which would
// otherwise concatenate a token JSON onto the same response body).
func TestHandleSubmitChallengeReturnsAfterFinalizeLoginError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, srv := newTestServer(ctx, t, func(c *Config) {
		c.Storage = &updateAuthRequestErrStorage{Storage: c.Storage}
	})
	defer httpServer.Close()

	web3Conn := storage.Connector{
		ID:              "web3-stub",
		Type:            "mockCallback",
		Name:            "Web3 Stub",
		ResourceVersion: "1",
	}
	if err := srv.storage.CreateConnector(ctx, web3Conn); err != nil {
		t.Fatalf("create connector: %v", err)
	}
	srv.mu.Lock()
	srv.connectors[web3Conn.ID] = Connector{
		ResourceVersion: web3Conn.ResourceVersion,
		Connector:       stubWeb3Connector{},
	}
	srv.mu.Unlock()

	if err := srv.storage.CreateClient(ctx, storage.Client{
		ID:           "test-client",
		Secret:       "secret",
		RedirectURIs: []string{"http://localhost/cb"},
	}); err != nil {
		t.Fatalf("create client: %v", err)
	}

	connData, err := json.Marshal(web3ConnectorData{
		Address: "0x0000000000000000000000000000000000000001",
		Nonce:   "test-nonce",
	})
	if err != nil {
		t.Fatalf("marshal connector data: %v", err)
	}

	authReq := storage.AuthRequest{
		ID:            storage.NewID(),
		ClientID:      "test-client",
		ConnectorID:   web3Conn.ID,
		ConnectorData: connData,
		Expiry:        srv.now().Add(10 * time.Minute),
		RedirectURI:   "http://localhost/cb",
		Scopes:        []string{"openid"},
		HMACKey:       []byte("test-hmac-key"),
	}
	if err := srv.storage.CreateAuthRequest(ctx, authReq); err != nil {
		t.Fatalf("create auth request: %v", err)
	}

	form := url.Values{}
	form.Set("state", authReq.ID)
	form.Set("signature", "0x0")
	form.Set("domain", "http://localhost/cb")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/web3/submit_challenge", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	srv.handleSubmitChallenge(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d (body=%q)", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}

	dec := json.NewDecoder(rr.Body)

	var errBody struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
	if err := dec.Decode(&errBody); err != nil {
		t.Fatalf("decode first JSON object: %v", err)
	}
	if errBody.Message != "Login failure." {
		t.Fatalf("expected message %q, got %q", "Login failure.", errBody.Message)
	}

	var trailing json.RawMessage
	if err := dec.Decode(&trailing); err != io.EOF {
		t.Fatalf("expected only one JSON object in body, but found a trailing object: err=%v body=%q", err, string(trailing))
	}
}
