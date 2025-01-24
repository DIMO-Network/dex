package server

import (
	"fmt"
	"log/slog"

	"github.com/dexidp/dex/storage"
)

func CreateCustomToken(s storage.Storage, logger *slog.Logger, payload []byte) (string, error) {
	keys, err := s.GetKeys()
	if err != nil {
		logger.Error("Failed to get keys", "error", err)
		return "", err
	}

	signingKey := keys.SigningKey
	if signingKey == nil {
		return "", fmt.Errorf("no key to sign payload with")
	}
	signingAlg, err := signatureAlgorithm(signingKey)
	if err != nil {
		return "", err
	}

	tk, err := signPayload(signingKey, signingAlg, payload)
	if err != nil {
		return "", err
	}
	return tk, nil
}
