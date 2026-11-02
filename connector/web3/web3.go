// Package web3 implements logging in through a web3.js-compatible wallet
package web3

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/dexidp/dex/connector"
)

type Config struct {
	InfuraID string `json:"infuraId"`
	RPCURL   string `json:"rpcUrl"`
}

func (c *Config) Open(id string, logger *slog.Logger) (connector.Connector, error) {
	w := &web3Connector{infuraID: c.InfuraID, logger: *logger}
	if c.RPCURL != "" {
		ethClient, err := createEthClient(c.RPCURL)
		if err != nil {
			return nil, err
		}
		w.ethClient = ethClient
	} else {
		logger.Warn("No RPC URL specified, contract signature validation will not be available.")
	}

	return w, nil
}

type web3Connector struct {
	infuraID  string
	ethClient bind.ContractBackend
	logger    slog.Logger
}

func (c *web3Connector) InfuraID() string {
	return c.infuraID
}

var erc1271magicValue = [4]byte{0x16, 0x26, 0xba, 0x7e}

// https://gist.github.com/dcb9/385631846097e1f59e3cba3b1d42f3ed#file-eth_sign_verify-go
func (c *web3Connector) Verify(address, msg, signedMsg string) (connector.Identity, error) {
	addrb := common.HexToAddress(address)

	msgHash := signHash([]byte(msg))

	identity, err := c.VerifyEOASignature(addrb, msgHash, signedMsg)
	if err != nil {
		return c.VerifyERC1271Signature(addrb, msgHash, signedMsg)
	}

	return identity, nil
}

func (c *web3Connector) VerifyEOASignature(addr common.Address, msgHash []byte, signedMsg string) (identity connector.Identity, err error) {
	signb, err := hexutil.Decode(signedMsg)
	if err != nil {
		return identity, fmt.Errorf("could not decode hex string of signed nonce: %v", err)
	}

	if len(signb) != 65 {
		return identity, fmt.Errorf("signature has length %d != 65", len(signb))
	}

	// This is the v parameter in the signature. Per the yellow paper, 27 means even and 28
	// means odd.
	if signb[64] == 27 || signb[64] == 28 {
		signb[64] -= 27
	} else if signb[64] != 0 && signb[64] != 1 {
		// We allow 0 or 1 because some non-conformant devices like Ledger use it.
		return identity, fmt.Errorf("v byte %d, not one of 0, 1, 27, or 28", signb[64])
	}

	pubKey, err := crypto.SigToPub(msgHash, signb)
	if err != nil {
		return c.VerifyERC1271Signature(addr, msgHash, signedMsg)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	// These are byte arrays, so this is okay to do.
	if recoveredAddr != addr {
		return identity, fmt.Errorf("hash was not signed by %s and not %s", recoveredAddr, addr)
	}

	identity.UserID = addr.Hex()
	identity.Username = addr.Hex()
	return identity, nil
}

func (c *web3Connector) VerifyERC1271Signature(contractAddress common.Address, hash []byte, signedMsg string) (identity connector.Identity, err error) {
	signature, err := hexutil.Decode(signedMsg)
	if err != nil {
		return identity, fmt.Errorf("could not decode hex string of signed nonce: %v", err)
	}

	if c.ethClient == nil {
		c.logger.Error("Eth client was not initialized successfully", "error", err)
		return identity, errors.New("can't attempt to validate signature, no Ethereum client available")
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

	result, err := ct.IsValidSignature(nil, msgHash, signature)
	if err != nil {
		return identity, fmt.Errorf("error occurred completing login %w", err)
	}

	if result != erc1271magicValue {
		return identity, fmt.Errorf("given address and address recovered from signed message do not match")
	}

	identity = connector.Identity{
		UserID:   contractAddress.Hex(),
		Username: contractAddress.Hex(),
	}
	if signer, ok := recoverSigner(hash, signature); ok {
		data, err := json.Marshal(connector.Web3SignerData{SignerAddress: signer.Hex()})
		if err != nil {
			return identity, fmt.Errorf("failed to encode signer data: %w", err)
		}
		identity.ConnectorData = data
	}
	return identity, nil
}

func signHash(data []byte) []byte {
	return accounts.TextHash(data)
}

// recoverSigner returns the EOA behind a plain 65-byte ECDSA signature over hash, the same
// recovery a developer license account performs before it calls isSigner. Smart-account
// signatures (other lengths or recovery ids) report false, so those logins carry no signer.
func recoverSigner(hash []byte, signature []byte) (common.Address, bool) {
	if len(signature) != 65 {
		return common.Address{}, false
	}
	sig := make([]byte, 65)
	copy(sig, signature)
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	if sig[64] > 1 {
		return common.Address{}, false
	}
	pub, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return common.Address{}, false
	}
	return crypto.PubkeyToAddress(*pub), true
}

func createEthClient(rpcURL string) (bind.ContractBackend, error) {
	return ethclient.Dial(rpcURL)
}
