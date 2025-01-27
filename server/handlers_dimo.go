package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	siwe "github.com/spruceid/siwe-go"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/storage"
)

type web3ConnectorData struct {
	Address string `json:"address"`
	Nonce   string `json:"nonce"`
}

var addressRegex = regexp.MustCompile("^0x[a-fA-F0-9]{40}$")

// Create an authorization request and nonce for a web3 login. Return state (the request id)
// and a random nonce.
func (s *Server) handleGenerateChallenge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Couldn't parse parameters.")
		return
	}

	r.Form.Set("redirect_uri", r.Form.Get("domain"))

	authReq, err := s.parseAuthorizationRequest(r)
	if err != nil {
		s.logger.Error("Failed to parse authorization request", "error", err)

		switch authErr := err.(type) {
		case *redirectedAuthErr:
			authErr.Handler().ServeHTTP(w, r)
		case *displayedAuthErr:
			s.renderError(r, w, authErr.Status, err.Error())
		default:
			panic("unsupported error type")
		}

		return
	}

	u, err := url.Parse(authReq.RedirectURI)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Invalid redirect URI.")
		return
	}

	rawAddr := r.Form.Get("address")
	if !common.IsHexAddress(rawAddr) {
		s.renderErrorJSON(w, http.StatusBadRequest, "Invalid Ethereum address.")
		return
	}
	addr := common.HexToAddress(rawAddr)

	connID := mux.Vars(r)["connector"]
	_, err = s.getConnector(connID)
	if err != nil {
		s.logger.Error("Failed to get connector", "error", err)
		s.renderError(r, w, http.StatusBadRequest, "Could not retrieve connector.")
		return
	}

	// Set the connector being used for the login.
	if authReq.ConnectorID != "" && authReq.ConnectorID != connID {
		s.logger.Error("Mismatched connector ID in auth request", "received",
			authReq.ConnectorID, "expected", connID)
		s.renderError(r, w, http.StatusBadRequest, "Bad connector ID")
		return
	}

	authReq.ConnectorID = connID

	nonce, err := generateNonce()
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "No source of randomness.")
		return
	}
	options := map[string]interface{}{
		"statement": fmt.Sprintf("%s is asking you sign in.", u.Hostname()),
	}

	siweMessage, err := siwe.InitMessage(s.issuerURL.Host, addr.Hex(), s.issuerURL.String(), nonce, options)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	challenge := siweMessage.String()

	authReq.ConnectorData, err = json.Marshal(web3ConnectorData{Address: addr.Hex(), Nonce: challenge})
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Failed to create auth request.")
		return
	}

	// Actually create the auth request
	authReq.Expiry = s.now().Add(s.authRequestsValidFor)
	if err := s.storage.CreateAuthRequest(r.Context(), *authReq); err != nil {
		s.logger.Error("Failed to create authorization request", "error", err)
		s.renderError(r, w, http.StatusInternalServerError, "Failed to connect to the database.")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	s.renderJSON(w, struct {
		State     string `json:"state"`
		Challenge string `json:"challenge"`
	}{authReq.ID, challenge})
}

func (s *Server) handleChallenge(w http.ResponseWriter, r *http.Request) {
	var nonceReq struct {
		Address string `json:"address"`
		State   string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&nonceReq); err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Could not parse request body JSON.")
		return
	}

	// Check that this is a valid Ethereum address and convert it to checksum form.
	mixAddr, err := common.NewMixedcaseAddressFromString(nonceReq.Address)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, fmt.Sprintf("Invalid Ethereum address %s.", nonceReq.Address))
	}
	if !mixAddr.ValidChecksum() {
		s.logger.Warn("Incoming address not checksummed", "address", mixAddr.Original())
		nonceReq.Address = mixAddr.Address().Hex()
	}

	authReq, err := s.storage.GetAuthRequest(nonceReq.State)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Requested resource does not exist.")
		return
	}

	u, err := url.Parse(authReq.RedirectURI)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Invalid redirect URI")
	}

	nonce, err := generateNonce()
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "No source of randomness.")
		return
	}
	options := map[string]interface{}{
		"statement": fmt.Sprintf("%s is asking you sign in.", u.Hostname()),
	}

	siweMessage, err := siwe.InitMessage(s.issuerURL.Host, nonceReq.Address, s.issuerURL.String(), nonce, options)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	challenge := siweMessage.String()

	bts, err := json.Marshal(web3ConnectorData{Address: nonceReq.Address, Nonce: challenge})
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Failed to create auth request.")
		return
	}
	s.storage.UpdateAuthRequest(authReq.ID, func(a storage.AuthRequest) (storage.AuthRequest, error) {
		a.ConnectorData = bts
		return a, nil
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	s.renderJSON(w, struct {
		Nonce string `json:"nonce"`
	}{challenge})
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

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	var verifyReq struct {
		Signed string `json:"signed"`
		State  string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&verifyReq); err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Could not parse request body JSON.")
		return
	}

	authReq, err := s.storage.GetAuthRequest(verifyReq.State)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Requested resource does not exist.")
		return
	}

	var data web3ConnectorData
	json.Unmarshal(authReq.ConnectorData, &data)

	conn, err := s.getConnector(authReq.ConnectorID)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Requested resource does not exist.")
		return
	}

	w3Conn, ok := conn.Connector.(connector.Web3Connector)
	if !ok {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Requested resource does not exist.")
		return
	}

	identity, err := w3Conn.Verify(data.Address, data.Nonce, verifyReq.Signed)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Could not verify signature.")
		return
	}

	redirectURL, canSkipApproval, err := s.finalizeLogin(r.Context(), identity, authReq, conn)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Login failure.")
	}

	if canSkipApproval {
		// We're overriding this. Always redirect.
		h := hmac.New(sha256.New, authReq.HMACKey)
		h.Write([]byte(authReq.ID))
		mac := h.Sum(nil)

		redirectURL = path.Join(s.issuerURL.Path, "/approval") + "?req=" + authReq.ID + "&hmac=" + base64.RawURLEncoding.EncodeToString(mac)
	}

	s.renderJSON(w, struct {
		Redirect string `json:"redirect"`
	}{redirectURL})
}

func (s *Server) handleVerifyDirect(w http.ResponseWriter, r *http.Request) {
	var verifyReq struct {
		Signed string `json:"signed"`
		State  string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&verifyReq); err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Could not parse request body JSON.")
		return
	}

	authReq, err := s.storage.GetAuthRequest(verifyReq.State)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Requested resource does not exist.")
		return
	}

	var data web3ConnectorData
	json.Unmarshal(authReq.ConnectorData, &data)

	conn, err := s.getConnector(authReq.ConnectorID)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Requested resource does not exist.")
		return
	}

	w3Conn, ok := conn.Connector.(connector.Web3Connector)
	if !ok {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Requested resource does not exist.")
		return
	}

	identity, err := w3Conn.Verify(data.Address, data.Nonce, verifyReq.Signed)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Could not verify signature.")
		return
	}

	_, _, err = s.finalizeLogin(r.Context(), identity, authReq, conn)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Login failure.")
	}

	// Need to pick up the changes made by finalizeLogin. This is pretty gross!
	authReq, err = s.storage.GetAuthRequest(verifyReq.State)
	if err != nil {
		s.logger.Error("Failed to get auth request", "error", err)
		s.renderError(r, w, http.StatusInternalServerError, "Database error.")
		return
	}

	if s.now().After(authReq.Expiry) {
		s.renderErrorJSON(w, http.StatusBadRequest, "User session has expired.")
		return
	}

	if err := s.storage.DeleteAuthRequest(authReq.ID); err != nil {
		if err != storage.ErrNotFound {
			s.logger.Error("Failed to delete authorization request", "error", err)
			s.renderErrorJSON(w, http.StatusInternalServerError, "Internal server error.")
		} else {
			s.renderErrorJSON(w, http.StatusBadRequest, "User session error.")
		}
		return
	}

	code := storage.AuthCode{
		ID:            storage.NewID(),
		ClientID:      authReq.ClientID,
		ConnectorID:   authReq.ConnectorID,
		Nonce:         authReq.Nonce,
		Scopes:        authReq.Scopes,
		Claims:        authReq.Claims,
		Expiry:        s.now().Add(time.Minute * 30),
		RedirectURI:   authReq.RedirectURI,
		ConnectorData: authReq.ConnectorData,
		PKCE:          authReq.PKCE,
	}
	if err := s.storage.CreateAuthCode(r.Context(), code); err != nil {
		s.logger.Error("Failed to create auth code", "error", err)
		s.renderError(r, w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	s.renderJSON(w, struct {
		Redirect string `json:"code"`
	}{code.ID})
}

// Handle the usual token request, except instead of the code we look for
// state (the auth request) and the sugnature.

func (s *Server) handleSubmitChallenge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Couldn't parse form.")
		return
	}

	r.PostForm.Set("redirect_uri", r.Form.Get("domain"))

	authReqID := r.PostFormValue("state")
	authReq, err := s.storage.GetAuthRequest(authReqID)
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Requested resource does not exist.")
		return
	}

	var data web3ConnectorData
	json.Unmarshal(authReq.ConnectorData, &data)

	conn, err := s.getConnector(authReq.ConnectorID)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Requested resource does not exist.")
		return
	}

	w3Conn, ok := conn.Connector.(connector.Web3Connector)
	if !ok {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Requested resource does not exist.")
		return
	}

	identity, err := w3Conn.Verify(data.Address, data.Nonce, r.PostFormValue("signature"))
	if err != nil {
		s.renderErrorJSON(w, http.StatusBadRequest, "Could not verify signature.")
		return
	}

	_, _, err = s.finalizeLogin(r.Context(), identity, authReq, conn)
	if err != nil {
		s.renderErrorJSON(w, http.StatusInternalServerError, "Login failure.")
	}

	// Need to pick up the changes made by finalizeLogin. This is pretty gross!
	authReq, err = s.storage.GetAuthRequest(authReqID)
	if err != nil {
		s.logger.Error("Failed to get auth request", "error", err)
		s.renderError(r, w, http.StatusInternalServerError, "Database error.")
		return
	}

	if s.now().After(authReq.Expiry) {
		s.renderErrorJSON(w, http.StatusBadRequest, "User session has expired.")
		return
	}

	if err := s.storage.DeleteAuthRequest(authReq.ID); err != nil {
		if err != storage.ErrNotFound {
			s.logger.Error("Failed to delete authorization request", "error", err)
			s.renderErrorJSON(w, http.StatusInternalServerError, "Internal server error.")
		} else {
			s.renderErrorJSON(w, http.StatusBadRequest, "User session error.")
		}
		return
	}

	code := storage.AuthCode{
		ID:            storage.NewID(),
		ClientID:      authReq.ClientID,
		ConnectorID:   authReq.ConnectorID,
		Nonce:         authReq.Nonce,
		Scopes:        authReq.Scopes,
		Claims:        authReq.Claims,
		Expiry:        s.now().Add(time.Minute * 30),
		RedirectURI:   authReq.RedirectURI,
		ConnectorData: authReq.ConnectorData,
		PKCE:          authReq.PKCE,
	}
	if err := s.storage.CreateAuthCode(r.Context(), code); err != nil {
		s.logger.Error("Failed to create auth code", "error", err)
		s.renderError(r, w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	r.PostForm.Set("code", code.ID)

	s.handleToken(w, r)
}

func (s *Server) handleCreateAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	authReq, err := s.parseAuthorizationRequest(r)
	if err != nil {
		s.logger.Error("Failed to parse authorization request", "error", err)

		switch authErr := err.(type) {
		case *redirectedAuthErr:
			authErr.Handler().ServeHTTP(w, r)
		case *displayedAuthErr:
			s.renderError(r, w, authErr.Status, err.Error())
		default:
			panic("unsupported error type")
		}

		return
	}

	connID := mux.Vars(r)["connector"]
	_, err = s.getConnector(connID)
	if err != nil {
		s.logger.Error("Failed to get connector", "error", err)
		s.renderError(r, w, http.StatusBadRequest, "Requested resource does not exist")
		return
	}

	// Set the connector being used for the login.
	if authReq.ConnectorID != "" && authReq.ConnectorID != connID {
		s.logger.Error("Mismatched connector ID in auth request", "got", authReq.ConnectorID, "expected", connID)
		s.renderError(r, w, http.StatusBadRequest, "Bad connector ID")
		return
	}

	authReq.ConnectorID = connID

	// Actually create the auth request
	authReq.Expiry = s.now().Add(s.authRequestsValidFor)
	if err := s.storage.CreateAuthRequest(r.Context(), *authReq); err != nil {
		s.logger.Error("Failed to create authorization request", "error", err)
		s.renderError(r, w, http.StatusInternalServerError, "Failed to connect to the database.")
		return
	}

	s.renderJSON(w, struct {
		State string `json:"state"`
	}{authReq.ID})
}

func (s *Server) renderJSON(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		s.logger.Error("Failed to write JSON response", "error", err)
	}
}

func (s *Server) renderErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		Status:  status,
		Message: message,
	}); err != nil {
		s.logger.Error("Failed to write JSON error response", "error", err)
	}
}
