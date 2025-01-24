package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dexidp/dex/api/v2"
)

// DIMO-specific functionality
func (d dexAPI) SignToken(ctx context.Context, req *api.SignTokenReq) (*api.SignTokenResp, error) {

	issuedAt := d.serverConfig.Now()
	expiry := issuedAt.Add(d.serverConfig.IDTokensValidFor)

	baseClaims := map[string]any{
		"sub": req.Subject,
		"iat": issuedAt.Unix(),
		"exp": expiry.Unix(),
		"iss": d.serverConfig.Issuer,
		"aud": req.Audience,
	}

	claims := req.CustomClaims.AsMap()

	// Base claims take priority.
	for k, v := range baseClaims {
		claims[k] = v
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("could not serialize claims: %v", err)
	}

	token, err := CreateCustomToken(d.s, d.logger, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %v", err)
	}

	return &api.SignTokenResp{
		Token: token,
	}, nil
}
