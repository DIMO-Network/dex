package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/storage"
)

var addressRegex = regexp.MustCompile("^0x[a-fA-F0-9]{40}$")

// DIMOClaims are the claims DIMO adds to every ID and access token.
type DIMOClaims struct {
	// ProviderID is the connector the user logged in with.
	ProviderID string `json:"provider_id"`
	// EthereumAddress is the user's wallet, for users whose ID is an address.
	EthereumAddress string `json:"ethereum_address,omitempty"`
	// SignerAddress is the developer-license signer (API key) that signed the web3
	// challenge, set on ERC-1271 developer-license logins.
	SignerAddress string `json:"signer_address,omitempty"`
}

// applyDIMOClaims fills tok's DIMO claims from the authorization.
func applyDIMOClaims(tok *IDTokenClaims, auth Authorization) {
	tok.ProviderID = auth.ConnectorID
	tok.SignerAddress = signerAddress(auth.ConnectorData)
	// This is a hack
	if slices.Contains(auth.Scopes, ScopeEmail) && addressRegex.MatchString(auth.Claims.UserID) {
		tok.EthereumAddress = auth.Claims.UserID
	}
}

// signerAddress returns, checksummed, the signer the web3 connector stored in its connector
// data (see connector.Web3SignerData). Connector data from any other connector, or without a
// well-formed address, yields "".
//
// Only grants that set Authorization.ConnectorData carry the claim, which among the web3
// flows means redeeming an auth code. Implicit and hybrid responses (authflow's response.go)
// mint tokens without it; production enables only the code flow (normalizeConfig defaults
// SupportedResponseTypes to [code]). Enabling either flow needs ConnectorData set there.
func signerAddress(connectorData []byte) string {
	if len(connectorData) == 0 {
		return ""
	}
	var data connector.Web3SignerData
	if err := json.Unmarshal(connectorData, &data); err != nil || !addressRegex.MatchString(data.SignerAddress) {
		return ""
	}
	return common.HexToAddress(data.SignerAddress).Hex()
}

// clientTokenLifetime returns how long client's tokens stay valid: the global setting, unless
// the client's IDTokenExpiry shortens it. It takes the client from the caller rather than
// storage because a grant signs after spending its code or refresh token, so a failed read
// here would lose them.
func (i *Issuer) clientTokenLifetime(ctx context.Context, client storage.Client) time.Duration {
	if client.IDTokenExpiry == "" {
		return i.idTokensValidFor
	}
	switch validFor, err := time.ParseDuration(client.IDTokenExpiry); {
	case err != nil:
		i.logger.ErrorContext(ctx, "client custom ID token expiry not a valid duration, this should never happen; using global",
			"client_id", client.ID, "id_expiry", client.IDTokenExpiry, "global_expiry", i.idTokensValidFor)
		return i.idTokensValidFor
	case validFor > i.idTokensValidFor:
		i.logger.ErrorContext(ctx, "client custom ID token expiry longer than the global setting; using the latter",
			"client_id", client.ID, "id_expiry", client.IDTokenExpiry, "global_expiry", i.idTokensValidFor)
		return i.idTokensValidFor
	default:
		return validFor
	}
}

// SignCustomToken signs a token for subject carrying customClaims. The base claims win over
// custom claims of the same name. It lasts the global ID token lifetime.
func (i *Issuer) SignCustomToken(ctx context.Context, subject string, audience []string, customClaims map[string]any) (string, error) {
	issuedAt := i.now()
	expiry := issuedAt.Add(i.idTokensValidFor)

	claims := make(map[string]any, len(customClaims)+5)
	for k, v := range customClaims {
		claims[k] = v
	}
	claims["sub"] = subject
	claims["iat"] = issuedAt.Unix()
	claims["exp"] = expiry.Unix()
	claims["iss"] = i.issuerURL.String()
	claims["aud"] = audience

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("could not serialize claims: %v", err)
	}

	token, err := i.signer.Sign(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %v", err)
	}
	return token, nil
}
