package internal

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/dexidp/dex/pkg/log"
	"github.com/dexidp/dex/storage"
	"gopkg.in/square/go-jose.v2"
)

type jwtHelper struct {
	key *jose.JSONWebKey
	alg jose.SignatureAlgorithm
}

// Determine the signature algorithm for a JWT.
func signatureAlgorithm(jwk *jose.JSONWebKey) (alg jose.SignatureAlgorithm, err error) {
	if jwk.Key == nil {
		return alg, errors.New("no signing key")
	}
	switch key := jwk.Key.(type) {
	case *rsa.PrivateKey:
		// Because OIDC mandates that we support RS256, we always return that
		// value. In the future, we might want to make this configurable on a
		// per client basis. For example allowing PS256 or ECDSA variants.
		//
		// See https://github.com/dexidp/dex/issues/692
		return jose.RS256, nil
	case *ecdsa.PrivateKey:
		// We don't actually support ECDSA keys yet, but they're tested for
		// in case we want to in the future.
		//
		// These values are prescribed depending on the ECDSA key type. We
		// can't return different values.
		switch key.Params() {
		case elliptic.P256().Params():
			return jose.ES256, nil
		case elliptic.P384().Params():
			return jose.ES384, nil
		case elliptic.P521().Params():
			return jose.ES512, nil
		default:
			return alg, errors.New("unsupported ecdsa curve")
		}
	default:
		return alg, fmt.Errorf("unsupported signing key type %T", key)
	}
}

func NewJwtHelper(s storage.Storage, logger log.Logger) (*jwtHelper, error) {
	keys, err := s.GetKeys()
	if err != nil {
		logger.Errorf("Failed to get keys: %v", err)
		return nil, err
	}

	signingKey := keys.SigningKey
	if signingKey == nil {
		return nil, fmt.Errorf("no key to sign payload with")
	}
	signingAlg, err := signatureAlgorithm(signingKey)
	if err != nil {
		return nil, err
	}

	return &jwtHelper{
		key: signingKey,
		alg: signingAlg,
	}, nil
}

func (j jwtHelper) SignPayload(payload []byte) (jws string, err error) {
	signingKey := jose.SigningKey{Key: j.key, Algorithm: j.alg}

	signer, err := jose.NewSigner(signingKey, &jose.SignerOptions{})
	if err != nil {
		return "", fmt.Errorf("new signer: %v", err)
	}
	signature, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("signing payload: %v", err)
	}
	return signature.CompactSerialize()
}