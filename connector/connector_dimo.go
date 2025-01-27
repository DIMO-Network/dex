// Package connector defines interfaces for federated identity strategies.
package connector

import "net/http"

// CallbackPOSTConnector is an interface implemented by connectors which use an OAuth
// style redirect flow via response_mode=form_post to determine user information.
type CallbackConnectorFormPOST interface {
	// The initial URL to redirect the user to.
	//
	// OAuth2 implementations should request different scopes from the upstream
	// identity provider based on the scopes requested by the downstream client.
	// For example, if the downstream client requests a refresh token from the
	// server, the connector should also request a token from the provider.
	//
	// Many identity providers have arbitrary restrictions on refresh tokens. For
	// example Google only allows a single refresh token per client/user/scopes
	// combination, and wont return a refresh token even if offline access is
	// requested if one has already been issues. There's no good general answer
	// for these kind of restrictions, and may require this package to become more
	// aware of the global set of user/connector interactions.
	LoginURL(s Scopes, callbackURL, state string) (string, error)

	// Handle the callback to the server and return an identity.
	HandlePOST(s Scopes, r *http.Request) (identity Identity, err error)
}

// A Web3Connector verifies ownership of an Ethereum account.
type Web3Connector interface {
	// Infura ID returns the configured Infura ID if one was provided, or else the
	// empty string.
	InfuraID() string

	// Verify checks that the given message was signed by the private key of the given
	// account.
	Verify(address, msg, signedMsg string) (identity Identity, err error)
}
