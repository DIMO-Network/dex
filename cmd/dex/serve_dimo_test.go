package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dexidp/dex/storage"
)

func TestValidateIDTokenExpiry(t *testing.T) {
	require.NoError(t, validateIDTokenExpiry(storage.Client{ID: "unset"}))
	require.NoError(t, validateIDTokenExpiry(storage.Client{ID: "valid", IDTokenExpiry: "1h"}))
	require.ErrorContains(t, validateIDTokenExpiry(storage.Client{ID: "invalid", IDTokenExpiry: "soon"}), `client "invalid"`)
}
