package apiserver

import (
	"context"

	"github.com/dexidp/dex/api/v2"
	"github.com/dexidp/dex/server/tokens"
)

// WithSignToken adds DIMO's SignToken method to a gRPC API server.
func WithSignToken(s api.DexServer, issuer *tokens.Issuer) api.DexServer {
	return signTokenAPI{DexServer: s, issuer: issuer}
}

type signTokenAPI struct {
	api.DexServer

	issuer *tokens.Issuer
}

// DIMO-specific functionality
func (d signTokenAPI) SignToken(ctx context.Context, req *api.SignTokenReq) (*api.SignTokenResp, error) {
	token, err := d.issuer.SignCustomToken(ctx, req.Subject, req.Audience, req.CustomClaims.AsMap())
	if err != nil {
		return nil, err
	}

	return &api.SignTokenResp{
		Token: token,
	}, nil
}
