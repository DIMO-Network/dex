// Package web3 implements logging in through a web3.js-compatible wallet
package web3

import (
	"bytes"
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
	w := &web3Connector{infuraID: c.InfuraID, logger: logger}
	return w, nil
}

type web3Connector struct {
	infuraID  string
	ethClient bind.ContractBackend
	logger    lg.Logger
}

func (c *web3Connector) InfuraID() string {
	return c.infuraID
}

func (c *web3Connector) SetEthClient(ethClient bind.ContractBackend) {
	c.ethClient = ethClient
}

// https://gist.github.com/dcb9/385631846097e1f59e3cba3b1d42f3ed#file-eth_sign_verify-go
func (c *web3Connector) Verify(address, msg, signedMsg string) (identity connector.Identity, err error) {
	addrb := common.HexToAddress(address)

	signb, err := hexutil.Decode(signedMsg)
	if err != nil {
		return identity, fmt.Errorf("could not decode hex string of signed nonce: %v", err)
	}

	msgHash := signHash([]byte(msg))

	// This is the v parameter in the signature. Per the yellow paper, 27 means even and 28
	// means odd.
	if signb[64] == 27 || signb[64] == 28 {
		signb[64] -= 27
	} else if signb[64] != 0 && signb[64] != 1 {
		// We allow 0 or 1 because some non-conformant devices like Ledger use it.
		return c.VerifyERC1271Signature(addrb, msgHash, signedMsg)
	}

	pubKey, err := crypto.SigToPub(msgHash, signb)
	if err != nil {
		return c.VerifyERC1271Signature(addrb, msgHash, signedMsg)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	// These are byte arrays, so this is okay to do.
	if recoveredAddr != addrb {
		return c.VerifyERC1271Signature(addrb, msgHash, signedMsg)
		// identity, fmt.Errorf("given address and address recovered from signed nonce do not match")
	}

	identity.UserID = address
	identity.Username = address
	return identity, nil
}

func (c *web3Connector) VerifyERC1271Signature(contractAddress common.Address, hash []byte, signedMsg string) (identity connector.Identity, err error) {
	signature, err := hexutil.Decode(signedMsg)
	if err != nil {
		return identity, fmt.Errorf("could not decode hex string of signed nonce: %v", err)
	}

	if c.ethClient == nil {
		c.logger.Errorf("Eth client was not initialized successfully %v", err)
		return identity, errors.New("error occurred completing authentication, please try again")
	}
	var msgHash [32]byte
	copy(msgHash[:], hash)

	// ContractLogin is just a simple interface with the signature for IsValidSignature
	/**
	 * function isValidSignature(bytes32 hash, bytes memory signature) external view returns (bytes4 magicValue);
	 */
	ct, err := NewErc1271(contractAddress, c.ethClient)
	if err != nil {
		return identity, fmt.Errorf("error occurred completing login %w", err)
	}

	isValid, err := ct.IsValidSignature(&bind.CallOpts{}, msgHash, signature)
	if err != nil {
		return identity, fmt.Errorf("error occurred completing login %w", err)
	}
	resultVal := common.BytesToAddress(isValid[:])
	truthyVal := common.HexToAddress("0x1626ba7e")

	if !bytes.Equal(truthyVal.Bytes(), resultVal.Bytes()) {
		return identity, fmt.Errorf("given address and address recovered from signed message do not match")
	}

	return connector.Identity{
		UserID:   contractAddress.Hex(),
		Username: contractAddress.Hex(),
	}, nil
}

func signHash(data []byte) []byte {
	return accounts.TextHash(data)
}
