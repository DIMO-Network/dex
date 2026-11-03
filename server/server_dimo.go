package server

import (
	"context"

	"github.com/dexidp/dex/server/mfa"
	"github.com/dexidp/dex/server/tokens"
)

// Issuer is the server's token issuer. The gRPC API needs it to sign DIMO's custom tokens.
func (s *Server) Issuer() *tokens.Issuer { return s.issuer }

// mfaChainResolver returns the MFA handler's chain resolution, for the login endpoints that
// have to decide without going through it.
func (s *Server) mfaChainResolver(c Config) func(ctx context.Context, clientID, connectorID string) ([]string, error) {
	h := &mfa.Handler{
		Storage:         s.storage,
		MFAProviders:    c.MFAProviders,
		DefaultMFAChain: c.DefaultMFAChain,
		Connectors:      s.connectors,
	}
	return h.ChainForClient
}
