package apiserver

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/dexidp/dex/api/v2"
	"github.com/dexidp/dex/server/signer"
	"github.com/dexidp/dex/server/tokens"
	"github.com/dexidp/dex/storage/memory"
)

func TestSignToken(t *testing.T) {
	logger := newLogger(t)
	store := memory.New(logger)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	sig, err := signer.NewMockSigner(key)
	require.NoError(t, err)
	issuerURL, err := url.Parse("https://auth.example.com")
	require.NoError(t, err)
	issuer := tokens.NewIssuer(store, sig, *issuerURL, time.Hour, time.Now, logger)

	// Served over gRPC, as cmd/dex registers it, so a missing registration fails here.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	serv := grpc.NewServer()
	api.RegisterDexServer(serv, WithSignToken(NewAPI(store, logger, "test", nil, nil, nil), issuer))
	go serv.Serve(l)
	t.Cleanup(serv.Stop)

	conn, err := grpc.NewClient(l.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	client := api.NewDexClient(conn)

	custom, err := structpb.NewStruct(map[string]any{
		"privilege_ids": []any{1.0, 4.0},
		"sub":           "spoofed",
		"iss":           "spoofed",
	})
	require.NoError(t, err)

	resp, err := client.SignToken(t.Context(), &api.SignTokenReq{
		Subject:      "0xabc",
		CustomClaims: custom,
		Audience:     []string{"dimo.zone"},
	})
	require.NoError(t, err)

	jws, err := jose.ParseSigned(resp.Token, []jose.SignatureAlgorithm{jose.RS256, jose.ES256})
	require.NoError(t, err)
	payload, err := jws.Verify(key.Public())
	require.NoError(t, err)

	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))

	// Base claims win over anything passed in as a custom claim.
	require.Equal(t, "0xabc", claims["sub"])
	require.Equal(t, issuerURL.String(), claims["iss"])
	require.Equal(t, []any{"dimo.zone"}, claims["aud"])
	require.Equal(t, []any{1.0, 4.0}, claims["privilege_ids"])
	require.Equal(t, time.Hour.Seconds(), claims["exp"].(float64)-claims["iat"].(float64))

	// The wrapper leaves the rest of the API alone.
	version, err := client.GetVersion(t.Context(), &api.VersionReq{})
	require.NoError(t, err)
	require.Equal(t, "test", version.Server)

	// A request without custom claims is fine.
	_, err = client.SignToken(t.Context(), &api.SignTokenReq{Subject: "0xabc"})
	require.NoError(t, err)
}
