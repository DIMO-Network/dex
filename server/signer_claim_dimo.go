package server

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dexidp/dex/connector"
)

// signerAddressKey carries the developer-license signer from an auth code to the tokens
// minted for it.
type signerAddressKey struct{}

// withSignerAddress returns ctx carrying, checksummed, the signer the web3 connector stored
// in the auth code's connector data (see connector.Web3SignerData). Connector data from any
// other connector, or without a well-formed address, leaves ctx unchanged.
//
// Only exchangeAuthCode calls this, so only the authorization-code flow carries the claim.
// Implicit and hybrid responses (handleApproval's responseTypeToken and responseTypeIDToken
// cases) mint tokens without it; production enables only the code flow (server.go defaults
// SupportedResponseTypes to [code]). Enabling either flow needs the same call there.
func withSignerAddress(ctx context.Context, connectorData []byte) context.Context {
	if len(connectorData) == 0 {
		return ctx
	}
	var data connector.Web3SignerData
	if err := json.Unmarshal(connectorData, &data); err != nil || !addressRegex.MatchString(data.SignerAddress) {
		return ctx
	}
	return context.WithValue(ctx, signerAddressKey{}, common.HexToAddress(data.SignerAddress).Hex())
}

// signerAddressFromContext returns the signer set by withSignerAddress, or "".
func signerAddressFromContext(ctx context.Context) string {
	signer, _ := ctx.Value(signerAddressKey{}).(string)
	return signer
}
