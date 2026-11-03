package mfa

import "context"

// ChainForClient returns the authenticators that apply to clientID's login through
// connectorID.
func (h *Handler) ChainForClient(ctx context.Context, clientID, connectorID string) ([]string, error) {
	return h.chainForClient(ctx, clientID, connectorID)
}
