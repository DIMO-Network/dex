// Package web3 implements logging in through a web3.js-compatible wallet
package web3

import (
	"errors"
	"fmt"
	"github.com/dexidp/dex/connector"
	lg "github.com/dexidp/dex/pkg/log"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Config struct {
	InfuraID string `json:"infuraId"`
	RpcURL   string `json:"rpcUrl"`
}

func (c *Config) Open(id string, logger lg.Logger) (connector.Connector, error) {
	return &web3Connector{infuraID: c.InfuraID, rpcUrl: c.RpcURL}, nil
}

type web3Connector struct {
	infuraID string
	rpcUrl   string
}

func (c *web3Connector) InfuraID() string {
	return c.infuraID
}

func (c *web3Connector) RpcURL() string {
	return c.rpcUrl
}

// https://gist.github.com/dcb9/385631846097e1f59e3cba3b1d42f3ed#file-eth_sign_verify-go
func (c *web3Connector) Verify(address, msg, signedMsg string, ethClient bind.ContractBackend) (identity connector.Identity, err error) {
	addrb := common.HexToAddress(address)

	signb, err := hexutil.Decode(signedMsg)
	if err != nil {
		return identity, fmt.Errorf("could not decode hex string of signed nonce: %v", err)
	}

	// This is the v parameter in the signature. Per the yellow paper, 27 means even and 28
	// means odd.
	if signb[64] == 27 || signb[64] == 28 {
		signb[64] -= 27
	} else if signb[64] != 0 && signb[64] != 1 {
		// We allow 0 or 1 because some non-conformant devices like Ledger use it.
		return c.VerifyERC1271Signature(addrb, msg, signedMsg, ethClient)
	}

	pubKey, err := crypto.SigToPub(signHash([]byte(msg)), signb)
	if err != nil {
		return c.VerifyERC1271Signature(addrb, msg, signedMsg, ethClient)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	// These are byte arrays, so this is okay to do.
	if recoveredAddr != addrb {
		return identity, fmt.Errorf("given address and address recovered from signed nonce do not match")
	}

	identity.UserID = address
	identity.Username = address
	return identity, nil
}

func (c *web3Connector) VerifyERC1271Signature(signer common.Address, msg, signedMsg string, ethClient bind.ContractBackend) (identity connector.Identity, err error) {
	hash, err := hexutil.Decode(msg)
	if err != nil {
		return identity, fmt.Errorf("could not decode hex string of signed nonce: %v", err)
	}
	var msgHash [32]byte
	copy(msgHash[:], hash)

	signature, err := hexutil.Decode(signedMsg)
	if err != nil {
		return identity, fmt.Errorf("could not decode signature of signed nonce: %v", err)
	}

	ct, err := NewContractLogin(signer, ethClient)
	if err != nil {
		return identity, errors.New("error occurred completing login")
		// fmt.Errorf("could not decode hex string of signed nonce: %v", err)
	}

	isValid, err := ct.IsValidSignature(&bind.CallOpts{}, msgHash, signature)
	if err != nil {
		return identity, fmt.Errorf("given address and address recovered from signed nonce do not match")
	}
	fmt.Printf("tx sent: %s", string(isValid[:]))

	return connector.Identity{}, nil
}

func signHash(data []byte) []byte {
	return accounts.TextHash(data)
}
