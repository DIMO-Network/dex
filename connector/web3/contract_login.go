// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package web3

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ContractLoginMetaData contains all meta data concerning the ContractLogin contract.
var ContractLoginMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"isValidSignature\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"magicValue\",\"type\":\"bytes4\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ContractLoginABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractLoginMetaData.ABI instead.
var ContractLoginABI = ContractLoginMetaData.ABI

// ContractLogin is an auto generated Go binding around an Ethereum contract.
type ContractLogin struct {
	ContractLoginCaller     // Read-only binding to the contract
	ContractLoginTransactor // Write-only binding to the contract
	ContractLoginFilterer   // Log filterer for contract events
}

// ContractLoginCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractLoginCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractLoginTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractLoginTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractLoginFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractLoginFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractLoginSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractLoginSession struct {
	Contract     *ContractLogin    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractLoginCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractLoginCallerSession struct {
	Contract *ContractLoginCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ContractLoginTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractLoginTransactorSession struct {
	Contract     *ContractLoginTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ContractLoginRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractLoginRaw struct {
	Contract *ContractLogin // Generic contract binding to access the raw methods on
}

// ContractLoginCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractLoginCallerRaw struct {
	Contract *ContractLoginCaller // Generic read-only contract binding to access the raw methods on
}

// ContractLoginTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractLoginTransactorRaw struct {
	Contract *ContractLoginTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractLogin creates a new instance of ContractLogin, bound to a specific deployed contract.
func NewContractLogin(address common.Address, backend bind.ContractBackend) (*ContractLogin, error) {
	contract, err := bindContractLogin(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractLogin{ContractLoginCaller: ContractLoginCaller{contract: contract}, ContractLoginTransactor: ContractLoginTransactor{contract: contract}, ContractLoginFilterer: ContractLoginFilterer{contract: contract}}, nil
}

// NewContractLoginCaller creates a new read-only instance of ContractLogin, bound to a specific deployed contract.
func NewContractLoginCaller(address common.Address, caller bind.ContractCaller) (*ContractLoginCaller, error) {
	contract, err := bindContractLogin(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractLoginCaller{contract: contract}, nil
}

// NewContractLoginTransactor creates a new write-only instance of ContractLogin, bound to a specific deployed contract.
func NewContractLoginTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractLoginTransactor, error) {
	contract, err := bindContractLogin(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractLoginTransactor{contract: contract}, nil
}

// NewContractLoginFilterer creates a new log filterer instance of ContractLogin, bound to a specific deployed contract.
func NewContractLoginFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractLoginFilterer, error) {
	contract, err := bindContractLogin(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractLoginFilterer{contract: contract}, nil
}

// bindContractLogin binds a generic wrapper to an already deployed contract.
func bindContractLogin(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractLoginMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractLogin *ContractLoginRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractLogin.Contract.ContractLoginCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractLogin *ContractLoginRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractLogin.Contract.ContractLoginTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractLogin *ContractLoginRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractLogin.Contract.ContractLoginTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractLogin *ContractLoginCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractLogin.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractLogin *ContractLoginTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractLogin.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractLogin *ContractLoginTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractLogin.Contract.contract.Transact(opts, method, params...)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 hash, bytes signature) view returns(bytes4 magicValue)
func (_ContractLogin *ContractLoginCaller) IsValidSignature(opts *bind.CallOpts, hash [32]byte, signature []byte) ([4]byte, error) {
	var out []interface{}
	err := _ContractLogin.contract.Call(opts, &out, "isValidSignature", hash, signature)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 hash, bytes signature) view returns(bytes4 magicValue)
func (_ContractLogin *ContractLoginSession) IsValidSignature(hash [32]byte, signature []byte) ([4]byte, error) {
	return _ContractLogin.Contract.IsValidSignature(&_ContractLogin.CallOpts, hash, signature)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 hash, bytes signature) view returns(bytes4 magicValue)
func (_ContractLogin *ContractLoginCallerSession) IsValidSignature(hash [32]byte, signature []byte) ([4]byte, error) {
	return _ContractLogin.Contract.IsValidSignature(&_ContractLogin.CallOpts, hash, signature)
}
