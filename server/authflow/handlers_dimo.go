package authflow

// handlers_dimo.go holds DIMO's web3 login: the wallet challenge and verify
// endpoints, plus the connector-specific steps login.go hands off to.

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	siwe "github.com/spruceid/siwe-go"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/server/connectors"
	"github.com/dexidp/dex/server/router"
	"github.com/dexidp/dex/storage"
)

// errLoginCompleted refuses a change to an auth request that already holds a login.
var errLoginCompleted = errors.New("login already completed")

type web3ConnectorData struct {
	Address string `json:"address"`
	Nonce   string `json:"nonce"`
}

// mountDIMO registers the web3 login endpoints. They answer cross-origin calls
// from wallets and the developer SDKs.
func (h *Handler) mountDIMO(m router.Mux) {
	m.HandleCORS("/auth/{connector}/generate_challenge", h.handleGenerateChallenge)
	m.HandleCORS("/auth/{connector}/submit_challenge", h.handleSubmitChallenge)
	m.HandleCORS("/auth/{connector}/challenge", h.handleChallenge)
	m.HandleCORS("/auth/{connector}/verify", h.handleVerify)
	m.HandleCORS("/auth/{connector}/verify_direct", h.handleVerifyDirect)
	m.HandleCORS("/auth/{connector}/init", h.handleCreateAuthorizationRequest)
}

// startFormPOSTLogin redirects to a connector that answers with response_mode=form_post.
func (h *Handler) startFormPOSTLogin(w http.ResponseWriter, r *http.Request, conn connector.CallbackConnectorFormPOST, scopes connector.Scopes, connID, authID string) {
	// Use the auth request ID as the "state" token.
	callbackURL, err := conn.LoginURL(scopes, h.IssuerURL.AbsURL("/callback"), authID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "connector returned error when creating callback", "connector_id", connID, "err", err)
		h.renderError(r, w, http.StatusInternalServerError, "Login error.")
		return
	}
	http.Redirect(w, r, callbackURL, http.StatusFound)
}

// renderWeb3Login renders the page that asks the browser wallet to sign in.
func (h *Handler) renderWeb3Login(w http.ResponseWriter, r *http.Request, conn connector.Web3Connector, connID, authID string) {
	challengeURL := url.URL{Path: h.IssuerURL.AbsPath("/auth", connID, "challenge")}
	verifyURL := url.URL{Path: h.IssuerURL.AbsPath("/auth", connID, "verify")}
	if err := h.Templates.Web3Login(r, w, challengeURL.String(), verifyURL.String(), authID, conn.InfuraID()); err != nil {
		h.Logger.ErrorContext(r.Context(), "server template error", "err", err)
	}
}

// parseConnectorAuthRequest parses the authorization request of a web3 endpoint
// and binds it to the connector in the path, applying the checks handleConnectorLogin
// makes before it does the same. On failure it writes the response and returns false.
func (h *Handler) parseConnectorAuthRequest(w http.ResponseWriter, r *http.Request) (*storage.AuthRequest, bool) {
	ctx := r.Context()
	authReq, _, err := h.parseAuthorizationRequest(r)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to parse authorization request", "err", err)

		switch authErr := err.(type) {
		case *redirectedAuthErr:
			authErr.Handler().ServeHTTP(w, r)
		case *displayedAuthErr:
			h.renderError(r, w, authErr.Status, err.Error())
		default:
			panic("unsupported error type")
		}

		return nil, false
	}

	connID, err := url.PathUnescape(mux.Vars(r)["connector"])
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to parse connector", "err", err)
		h.renderError(r, w, http.StatusBadRequest, "Requested resource does not exist")
		return nil, false
	}

	client, authErr := h.getClientWithAuthError(ctx, authReq.ClientID)
	if authErr != nil {
		h.renderError(r, w, authErr.Status, authErr.Error())
		return nil, false
	}
	if !connectors.ConnectorAllowed(client.AllowedConnectors, connID) {
		h.Logger.ErrorContext(ctx, "connector not allowed for client", "connector_id", connID, "client_id", authReq.ClientID)
		h.renderErrorJSON(w, http.StatusForbidden, "Connector not allowed for this client.")
		return nil, false
	}

	conn, err := h.Connectors.Get(ctx, connID)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to get connector", "err", err)
		h.renderError(r, w, http.StatusBadRequest, "Could not retrieve connector.")
		return nil, false
	}

	if _, ok := conn.Connector.(connector.Web3Connector); !ok {
		h.Logger.ErrorContext(ctx, "web3 login requested on another connector", "connector_id", connID)
		h.renderErrorJSON(w, http.StatusBadRequest, "Requested connector does not support this login.")
		return nil, false
	}

	grantType := h.grantTypeFromAuthRequest(r)
	if !connectors.GrantTypeAllowed(conn.GrantTypes, grantType) {
		h.Logger.ErrorContext(ctx, "connector does not allow requested grant type", "connector_id", connID, "grant_type", grantType)
		h.renderErrorJSON(w, http.StatusBadRequest, "Requested connector does not support this grant type.")
		return nil, false
	}

	// Set the connector being used for the login.
	if authReq.ConnectorID != "" && authReq.ConnectorID != connID {
		h.Logger.ErrorContext(ctx, "mismatched connector ID in auth request",
			"auth_request_connector_id", authReq.ConnectorID, "connector_id", connID)
		h.renderError(r, w, http.StatusBadRequest, "Bad connector ID")
		return nil, false
	}

	authReq.ConnectorID = connID
	return authReq, true
}

// Create an authorization request and nonce for a web3 login. Return state (the request id)
// and a random nonce.
func (h *Handler) handleGenerateChallenge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Couldn't parse parameters.")
		return
	}

	r.Form.Set("redirect_uri", r.Form.Get("domain"))

	authReq, ok := h.parseConnectorAuthRequest(w, r)
	if !ok {
		return
	}

	u, err := url.Parse(authReq.RedirectURI)
	if err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Invalid redirect URI.")
		return
	}

	rawAddr := r.Form.Get("address")
	if !common.IsHexAddress(rawAddr) {
		h.renderErrorJSON(w, http.StatusBadRequest, "Invalid Ethereum address.")
		return
	}
	addr := common.HexToAddress(rawAddr)

	challenge, err := h.newChallenge(u, addr.Hex())
	if err != nil {
		h.renderErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	authReq.ConnectorData, err = json.Marshal(web3ConnectorData{Address: addr.Hex(), Nonce: challenge})
	if err != nil {
		h.renderErrorJSON(w, http.StatusInternalServerError, "Failed to create auth request.")
		return
	}

	// Actually create the auth request
	authReq.Expiry = h.Now().Add(h.AuthRequestsValidFor)
	if err := h.Storage.CreateAuthRequest(r.Context(), *authReq); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to create authorization request", "err", err)
		h.renderError(r, w, http.StatusInternalServerError, "Failed to connect to the database.")
		return
	}

	h.renderJSON(w, struct {
		State     string `json:"state"`
		Challenge string `json:"challenge"`
	}{authReq.ID, challenge})
}

// openWeb3Login loads the auth request state for a web3 login still in progress on the
// connector in the path, and returns it with that connector. A logged-in request holds the
// identity the /auth dispatcher and the token endpoint will use, so nothing here may change
// it. On failure it writes the response and returns false.
func (h *Handler) openWeb3Login(w http.ResponseWriter, r *http.Request, state string) (storage.AuthRequest, connector.Web3Connector, bool) {
	ctx := r.Context()
	authReq, err := h.Storage.GetAuthRequest(ctx, state)
	if err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Requested resource does not exist.")
		return storage.AuthRequest{}, nil, false
	}

	if connID, err := url.PathUnescape(mux.Vars(r)["connector"]); err != nil || connID != authReq.ConnectorID {
		h.renderErrorJSON(w, http.StatusBadRequest, "Bad connector ID")
		return storage.AuthRequest{}, nil, false
	}

	if authReq.LoggedIn {
		h.renderErrorJSON(w, http.StatusBadRequest, "Login already completed.")
		return storage.AuthRequest{}, nil, false
	}

	if h.Now().After(authReq.Expiry) {
		h.renderErrorJSON(w, http.StatusBadRequest, "User session has expired.")
		return storage.AuthRequest{}, nil, false
	}

	conn, err := h.Connectors.Get(ctx, authReq.ConnectorID)
	if err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Requested resource does not exist.")
		return storage.AuthRequest{}, nil, false
	}

	w3Conn, ok := conn.Connector.(connector.Web3Connector)
	if !ok {
		h.renderErrorJSON(w, http.StatusBadRequest, "Requested resource does not exist.")
		return storage.AuthRequest{}, nil, false
	}
	return authReq, w3Conn, true
}

func (h *Handler) handleChallenge(w http.ResponseWriter, r *http.Request) {
	var nonceReq struct {
		Address string `json:"address"`
		State   string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&nonceReq); err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Could not parse request body JSON.")
		return
	}

	// Check that this is a valid Ethereum address and convert it to checksum form.
	mixAddr, err := common.NewMixedcaseAddressFromString(nonceReq.Address)
	if err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, fmt.Sprintf("Invalid Ethereum address %s.", nonceReq.Address))
		return
	}
	if !mixAddr.ValidChecksum() {
		h.Logger.WarnContext(r.Context(), "incoming address not checksummed", "address", mixAddr.Original())
		nonceReq.Address = mixAddr.Address().Hex()
	}

	authReq, _, ok := h.openWeb3Login(w, r, nonceReq.State)
	if !ok {
		return
	}

	u, err := url.Parse(authReq.RedirectURI)
	if err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Invalid redirect URI")
		return
	}

	challenge, err := h.newChallenge(u, nonceReq.Address)
	if err != nil {
		h.renderErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	bts, err := json.Marshal(web3ConnectorData{Address: nonceReq.Address, Nonce: challenge})
	if err != nil {
		h.renderErrorJSON(w, http.StatusInternalServerError, "Failed to create auth request.")
		return
	}
	// A login may have finished since openWeb3Login: check again where the write happens.
	err = h.Storage.UpdateAuthRequest(r.Context(), authReq.ID, func(a storage.AuthRequest) (storage.AuthRequest, error) {
		if a.LoggedIn {
			return a, errLoginCompleted
		}
		a.ConnectorData = bts
		return a, nil
	})
	switch {
	case errors.Is(err, errLoginCompleted):
		h.renderErrorJSON(w, http.StatusBadRequest, "Login already completed.")
		return
	case err != nil:
		h.Logger.ErrorContext(r.Context(), "failed to store challenge", "err", err)
		h.renderErrorJSON(w, http.StatusInternalServerError, "Failed to update auth request.")
		return
	}

	h.renderJSON(w, struct {
		Nonce string `json:"nonce"`
	}{challenge})
}

// newChallenge builds the SIWE message address has to sign to log in to redirectURI's host.
func (h *Handler) newChallenge(redirectURI *url.URL, address string) (string, error) {
	nonce, err := generateNonce()
	if err != nil {
		return "", fmt.Errorf("no source of randomness: %w", err)
	}
	options := map[string]interface{}{
		"statement": fmt.Sprintf("%s is asking you sign in.", redirectURI.Hostname()),
	}

	siweMessage, err := siwe.InitMessage(h.IssuerURL.Host, address, h.IssuerURL.String(), nonce, options)
	if err != nil {
		return "", err
	}
	return siweMessage.String(), nil
}

func generateNonce() (string, error) {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphabetSize := big.NewInt(int64(len(alphabet)))
	b := make([]byte, 30)
	for i := range b {
		c, err := rand.Int(rand.Reader, alphabetSize)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[c.Int64()]
	}
	return string(b), nil
}

// verifySignature checks signed against the challenge stored on the auth request state
// and finalizes the login. On failure it writes the response and returns false.
func (h *Handler) verifySignature(w http.ResponseWriter, r *http.Request, state, signed string) (storage.AuthRequest, bool) {
	ctx := r.Context()
	authReq, w3Conn, ok := h.openWeb3Login(w, r, state)
	if !ok {
		return storage.AuthRequest{}, false
	}

	var data web3ConnectorData
	json.Unmarshal(authReq.ConnectorData, &data)

	identity, err := w3Conn.Verify(data.Address, data.Nonce, signed)
	if err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Could not verify signature.")
		return storage.AuthRequest{}, false
	}

	authReq, err = h.finalizeLogin(ctx, identity, authReq, w3Conn)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to finalize login", "err", err)
		h.renderErrorJSON(w, http.StatusInternalServerError, "Login failure.")
		return storage.AuthRequest{}, false
	}
	return authReq, true
}

func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request) {
	var verifyReq struct {
		Signed string `json:"signed"`
		State  string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&verifyReq); err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Could not parse request body JSON.")
		return
	}

	authReq, ok := h.verifySignature(w, r, verifyReq.State, verifyReq.Signed)
	if !ok {
		return
	}

	// As in the connector callback, a wallet login has no remember_me checkbox.
	rememberMe := h.Sessions.RememberMeDefault()
	if err := h.Sessions.CreateOrUpdateAuthSession(r.Context(), r, w, authReq, rememberMe != nil && *rememberMe); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to create/update auth session", "err", err)
	}

	// The browser goes back to the /auth dispatcher, which runs MFA and consent
	// if they apply and then issues the response.
	h.renderJSON(w, struct {
		Redirect string `json:"redirect"`
	}{h.buildContinueURL(authReq)})
}

func (h *Handler) handleVerifyDirect(w http.ResponseWriter, r *http.Request) {
	var verifyReq struct {
		Signed string `json:"signed"`
		State  string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&verifyReq); err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Could not parse request body JSON.")
		return
	}

	code, ok := h.issueDirectCode(w, r, verifyReq.State, verifyReq.Signed)
	if !ok {
		return
	}

	h.renderJSON(w, struct {
		Redirect string `json:"code"`
	}{code.ID})
}

// Handle the usual token request, except instead of the code we look for
// state (the auth request) and the sugnature.

func (h *Handler) handleSubmitChallenge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderErrorJSON(w, http.StatusBadRequest, "Couldn't parse form.")
		return
	}

	r.PostForm.Set("redirect_uri", r.Form.Get("domain"))

	code, ok := h.issueDirectCode(w, r, r.PostFormValue("state"), r.PostFormValue("signature"))
	if !ok {
		return
	}

	r.PostForm.Set("code", code.ID)

	h.TokenEndpoint(w, r)
}

// issueDirectCode verifies the signature and mints the auth code straight away, for the
// clients that log in without a browser: there is no redirect back through the /auth
// dispatcher, so no consent screen, and so no browser session on the code either. On
// failure it writes the response and returns false.
func (h *Handler) issueDirectCode(w http.ResponseWriter, r *http.Request, state, signed string) (storage.AuthCode, bool) {
	ctx := r.Context()
	authReq, ok := h.verifySignature(w, r, state, signed)
	if !ok {
		return storage.AuthCode{}, false
	}

	// MFA only runs on the dispatcher's path. Refuse rather than skip it.
	if h.MFAEnabled {
		chain, err := h.MFAChain(ctx, authReq.ClientID, authReq.ConnectorID)
		if err != nil {
			h.Logger.ErrorContext(ctx, "failed to determine MFA requirement", "err", err)
			h.renderErrorJSON(w, http.StatusInternalServerError, "Internal server error.")
			return storage.AuthCode{}, false
		}
		if len(chain) > 0 {
			h.renderErrorJSON(w, http.StatusForbidden, "This client requires MFA, which this login does not support.")
			return storage.AuthCode{}, false
		}
	}

	if h.Now().After(authReq.Expiry) {
		h.renderErrorJSON(w, http.StatusBadRequest, "User session has expired.")
		return storage.AuthCode{}, false
	}

	if err := h.Storage.DeleteAuthRequest(ctx, authReq.ID); err != nil {
		if err != storage.ErrNotFound {
			h.Logger.ErrorContext(ctx, "failed to delete authorization request", "err", err)
			h.renderErrorJSON(w, http.StatusInternalServerError, "Internal server error.")
		} else {
			h.renderErrorJSON(w, http.StatusBadRequest, "User session error.")
		}
		return storage.AuthCode{}, false
	}

	code := storage.AuthCode{
		ID:            storage.NewID(),
		ClientID:      authReq.ClientID,
		ConnectorID:   authReq.ConnectorID,
		Nonce:         authReq.Nonce,
		Scopes:        authReq.Scopes,
		Claims:        authReq.Claims,
		Expiry:        h.Now().Add(time.Minute * 30),
		RedirectURI:   authReq.RedirectURI,
		ConnectorData: authReq.ConnectorData,
		PKCE:          authReq.PKCE,
		AuthTime:      authReq.AuthTime,
	}
	if err := h.Storage.CreateAuthCode(ctx, code); err != nil {
		h.Logger.ErrorContext(ctx, "failed to create auth code", "err", err)
		h.renderError(r, w, http.StatusInternalServerError, "Internal server error.")
		return storage.AuthCode{}, false
	}
	return code, true
}

func (h *Handler) handleCreateAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	authReq, ok := h.parseConnectorAuthRequest(w, r)
	if !ok {
		return
	}

	// Actually create the auth request
	authReq.Expiry = h.Now().Add(h.AuthRequestsValidFor)
	if err := h.Storage.CreateAuthRequest(r.Context(), *authReq); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to create authorization request", "err", err)
		h.renderError(r, w, http.StatusInternalServerError, "Failed to connect to the database.")
		return
	}

	h.renderJSON(w, struct {
		State string `json:"state"`
	}{authReq.ID})
}

// resumableSession returns the session a login on conn may resume. A web3 login never
// resumes one: each signature proves the signer is still a signer, and the token carries
// that signer, while a resumed session has neither.
func (h *Handler) resumableSession(ctx context.Context, w http.ResponseWriter, r *http.Request, authReq *storage.AuthRequest, conn connector.Connector) *storage.AuthSession {
	if _, ok := conn.(connector.Web3Connector); ok {
		return nil
	}
	return h.Sessions.ValidAuthSession(ctx, w, r, authReq)
}

// tokenClient returns the client to sign clientID's implicit and hybrid tokens with, which
// has to carry its IDTokenExpiry. Only static clients set that and those always load, so a
// failed load means the client has no override.
func (h *Handler) tokenClient(ctx context.Context, clientID string) storage.Client {
	client, err := h.Storage.GetClient(ctx, clientID)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to get client", "client_id", clientID, "err", err)
	}
	return storage.Client{ID: clientID, IDTokenExpiry: client.IDTokenExpiry}
}

func (h *Handler) renderJSON(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		h.Logger.Error("failed to write JSON response", "err", err)
	}
}

func (h *Handler) renderErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		Status:  status,
		Message: message,
	}); err != nil {
		h.Logger.Error("failed to write JSON error response", "err", err)
	}
}
