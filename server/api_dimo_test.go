package server

import (
	"encoding/json"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/dexidp/dex/api/v2"
)

func TestSignToken(t *testing.T) {
	httpServer, s := newTestServer(t, nil)
	defer httpServer.Close()

	custom, err := structpb.NewStruct(map[string]any{
		"privilege_ids": []any{1.0, 4.0},
		"sub":           "spoofed",
		"iss":           "spoofed",
	})
	require.NoError(t, err)

	resp, err := NewAPI(s.storage, s.logger, "test", s).SignToken(t.Context(), &api.SignTokenReq{
		Subject:      "0xabc",
		CustomClaims: custom,
		Audience:     []string{"dimo.zone"},
	})
	require.NoError(t, err)

	jws, err := jose.ParseSigned(resp.Token, []jose.SignatureAlgorithm{jose.RS256, jose.ES256})
	require.NoError(t, err)
	payload, err := jws.Verify(testKey.Public())
	require.NoError(t, err)

	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))

	// Base claims win over anything passed in as a custom claim.
	require.Equal(t, "0xabc", claims["sub"])
	require.Equal(t, s.issuerURL.String(), claims["iss"])
	require.Equal(t, []any{"dimo.zone"}, claims["aud"])
	require.Equal(t, []any{1.0, 4.0}, claims["privilege_ids"])
	require.Equal(t, s.idTokensValidFor.Seconds(), claims["exp"].(float64)-claims["iat"].(float64))
}
