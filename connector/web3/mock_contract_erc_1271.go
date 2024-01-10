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

// ContractAuthMetaData contains all meta data concerning the ContractAuth contract.
var ContractAuthMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes4\",\"name\":\"isValid\",\"type\":\"bytes4\"}],\"name\":\"ValidatedSigner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"isValidSignature\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610898806100606000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80631626ba7e14610030575b600080fd5b61004a60048036038101906100459190610592565b610060565b604051610057919061062d565b60405180910390f35b6000806100b18585858080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505061022d565b905060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff160361019857631626ba7e60e01b8173ffffffffffffffffffffffffffffffffffffffff1660008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167fb0a1cb3375368319eda37d9ea023fe4d12b0185fed7e4022ff1729b4bf40b40b60405160405180910390a4631626ba7e60e01b915050610226565b63ffffffff60e01b8173ffffffffffffffffffffffffffffffffffffffff1660008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167fb0a1cb3375368319eda37d9ea023fe4d12b0185fed7e4022ff1729b4bf40b40b60405160405180910390a463ffffffff60e01b9150505b9392505050565b600080600061023c8585610254565b91509150610249816102a5565b819250505092915050565b60008060418351036102955760008060006020860151925060408601519150606086015160001a90506102898782858561040b565b9450945050505061029e565b60006002915091505b9250929050565b600060048111156102b9576102b8610648565b5b8160048111156102cc576102cb610648565b5b031561040857600160048111156102e6576102e5610648565b5b8160048111156102f9576102f8610648565b5b03610339576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610330906106d4565b60405180910390fd5b6002600481111561034d5761034c610648565b5b8160048111156103605761035f610648565b5b036103a0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161039790610740565b60405180910390fd5b600360048111156103b4576103b3610648565b5b8160048111156103c7576103c6610648565b5b03610407576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103fe906107d2565b60405180910390fd5b5b50565b6000807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08360001c11156104465760006003915091506104e4565b60006001878787876040516000815260200160405260405161046b949392919061081d565b6020604051602081039080840390855afa15801561048d573d6000803e3d6000fd5b505050602060405103519050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16036104db576000600192509250506104e4565b80600092509250505b94509492505050565b600080fd5b600080fd5b6000819050919050565b61050a816104f7565b811461051557600080fd5b50565b60008135905061052781610501565b92915050565b600080fd5b600080fd5b600080fd5b60008083601f8401126105525761055161052d565b5b8235905067ffffffffffffffff81111561056f5761056e610532565b5b60208301915083600182028301111561058b5761058a610537565b5b9250929050565b6000806000604084860312156105ab576105aa6104ed565b5b60006105b986828701610518565b935050602084013567ffffffffffffffff8111156105da576105d96104f2565b5b6105e68682870161053c565b92509250509250925092565b60007fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b610627816105f2565b82525050565b6000602082019050610642600083018461061e565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b600082825260208201905092915050565b7f45434453413a20696e76616c6964207369676e61747572650000000000000000600082015250565b60006106be601883610677565b91506106c982610688565b602082019050919050565b600060208201905081810360008301526106ed816106b1565b9050919050565b7f45434453413a20696e76616c6964207369676e6174757265206c656e67746800600082015250565b600061072a601f83610677565b9150610735826106f4565b602082019050919050565b600060208201905081810360008301526107598161071d565b9050919050565b7f45434453413a20696e76616c6964207369676e6174757265202773272076616c60008201527f7565000000000000000000000000000000000000000000000000000000000000602082015250565b60006107bc602283610677565b91506107c782610760565b604082019050919050565b600060208201905081810360008301526107eb816107af565b9050919050565b6107fb816104f7565b82525050565b600060ff82169050919050565b61081781610801565b82525050565b600060808201905061083260008301876107f2565b61083f602083018661080e565b61084c60408301856107f2565b61085960608301846107f2565b9594505050505056fea264697066735822122007cd41d95174e9782fbf969ae27e7b2ef0324555207d9371ec790f251839d8ef64736f6c63430008140033",
}

// ContractAuthABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractAuthMetaData.ABI instead.
var ContractAuthABI = ContractAuthMetaData.ABI

// ContractAuthBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractAuthMetaData.Bin instead.
var ContractAuthBin = ContractAuthMetaData.Bin

// DeployContractAuth deploys a new Ethereum contract, binding an instance of ContractAuth to it.
func DeployContractAuth(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ContractAuth, error) {
	parsed, err := ContractAuthMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractAuthBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractAuth{ContractAuthCaller: ContractAuthCaller{contract: contract}, ContractAuthTransactor: ContractAuthTransactor{contract: contract}, ContractAuthFilterer: ContractAuthFilterer{contract: contract}}, nil
}

// ContractAuth is an auto generated Go binding around an Ethereum contract.
type ContractAuth struct {
	ContractAuthCaller     // Read-only binding to the contract
	ContractAuthTransactor // Write-only binding to the contract
	ContractAuthFilterer   // Log filterer for contract events
}

// ContractAuthCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractAuthCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractAuthTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractAuthTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractAuthFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractAuthFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractAuthSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractAuthSession struct {
	Contract     *ContractAuth     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractAuthCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractAuthCallerSession struct {
	Contract *ContractAuthCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// ContractAuthTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractAuthTransactorSession struct {
	Contract     *ContractAuthTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ContractAuthRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractAuthRaw struct {
	Contract *ContractAuth // Generic contract binding to access the raw methods on
}

// ContractAuthCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractAuthCallerRaw struct {
	Contract *ContractAuthCaller // Generic read-only contract binding to access the raw methods on
}

// ContractAuthTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractAuthTransactorRaw struct {
	Contract *ContractAuthTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractAuth creates a new instance of ContractAuth, bound to a specific deployed contract.
func NewContractAuth(address common.Address, backend bind.ContractBackend) (*ContractAuth, error) {
	contract, err := bindContractAuth(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractAuth{ContractAuthCaller: ContractAuthCaller{contract: contract}, ContractAuthTransactor: ContractAuthTransactor{contract: contract}, ContractAuthFilterer: ContractAuthFilterer{contract: contract}}, nil
}

// NewContractAuthCaller creates a new read-only instance of ContractAuth, bound to a specific deployed contract.
func NewContractAuthCaller(address common.Address, caller bind.ContractCaller) (*ContractAuthCaller, error) {
	contract, err := bindContractAuth(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractAuthCaller{contract: contract}, nil
}

// NewContractAuthTransactor creates a new write-only instance of ContractAuth, bound to a specific deployed contract.
func NewContractAuthTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractAuthTransactor, error) {
	contract, err := bindContractAuth(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractAuthTransactor{contract: contract}, nil
}

// NewContractAuthFilterer creates a new log filterer instance of ContractAuth, bound to a specific deployed contract.
func NewContractAuthFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractAuthFilterer, error) {
	contract, err := bindContractAuth(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractAuthFilterer{contract: contract}, nil
}

// bindContractAuth binds a generic wrapper to an already deployed contract.
func bindContractAuth(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractAuthMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractAuth *ContractAuthRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractAuth.Contract.ContractAuthCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractAuth *ContractAuthRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractAuth.Contract.ContractAuthTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractAuth *ContractAuthRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractAuth.Contract.ContractAuthTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractAuth *ContractAuthCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractAuth.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractAuth *ContractAuthTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractAuth.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractAuth *ContractAuthTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractAuth.Contract.contract.Transact(opts, method, params...)
}

// IsValidSignature is a paid mutator transaction binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _hash, bytes _signature) returns(bytes4)
func (_ContractAuth *ContractAuthTransactor) IsValidSignature(opts *bind.TransactOpts, _hash [32]byte, _signature []byte) (*types.Transaction, error) {
	return _ContractAuth.contract.Transact(opts, "isValidSignature", _hash, _signature)
}

// IsValidSignature is a paid mutator transaction binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _hash, bytes _signature) returns(bytes4)
func (_ContractAuth *ContractAuthSession) IsValidSignature(_hash [32]byte, _signature []byte) (*types.Transaction, error) {
	return _ContractAuth.Contract.IsValidSignature(&_ContractAuth.TransactOpts, _hash, _signature)
}

// IsValidSignature is a paid mutator transaction binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _hash, bytes _signature) returns(bytes4)
func (_ContractAuth *ContractAuthTransactorSession) IsValidSignature(_hash [32]byte, _signature []byte) (*types.Transaction, error) {
	return _ContractAuth.Contract.IsValidSignature(&_ContractAuth.TransactOpts, _hash, _signature)
}

// ContractAuthValidatedSignerIterator is returned from FilterValidatedSigner and is used to iterate over the raw logs and unpacked data for ValidatedSigner events raised by the ContractAuth contract.
type ContractAuthValidatedSignerIterator struct {
	Event *ContractAuthValidatedSigner // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractAuthValidatedSignerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractAuthValidatedSigner)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractAuthValidatedSigner)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractAuthValidatedSignerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractAuthValidatedSignerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractAuthValidatedSigner represents a ValidatedSigner event raised by the ContractAuth contract.
type ContractAuthValidatedSigner struct {
	Owner   common.Address
	Signer  common.Address
	IsValid [4]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterValidatedSigner is a free log retrieval operation binding the contract event 0xb0a1cb3375368319eda37d9ea023fe4d12b0185fed7e4022ff1729b4bf40b40b.
//
// Solidity: event ValidatedSigner(address indexed owner, address indexed signer, bytes4 indexed isValid)
func (_ContractAuth *ContractAuthFilterer) FilterValidatedSigner(opts *bind.FilterOpts, owner []common.Address, signer []common.Address, isValid [][4]byte) (*ContractAuthValidatedSignerIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}
	var isValidRule []interface{}
	for _, isValidItem := range isValid {
		isValidRule = append(isValidRule, isValidItem)
	}

	logs, sub, err := _ContractAuth.contract.FilterLogs(opts, "ValidatedSigner", ownerRule, signerRule, isValidRule)
	if err != nil {
		return nil, err
	}
	return &ContractAuthValidatedSignerIterator{contract: _ContractAuth.contract, event: "ValidatedSigner", logs: logs, sub: sub}, nil
}

// WatchValidatedSigner is a free log subscription operation binding the contract event 0xb0a1cb3375368319eda37d9ea023fe4d12b0185fed7e4022ff1729b4bf40b40b.
//
// Solidity: event ValidatedSigner(address indexed owner, address indexed signer, bytes4 indexed isValid)
func (_ContractAuth *ContractAuthFilterer) WatchValidatedSigner(opts *bind.WatchOpts, sink chan<- *ContractAuthValidatedSigner, owner []common.Address, signer []common.Address, isValid [][4]byte) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}
	var isValidRule []interface{}
	for _, isValidItem := range isValid {
		isValidRule = append(isValidRule, isValidItem)
	}

	logs, sub, err := _ContractAuth.contract.WatchLogs(opts, "ValidatedSigner", ownerRule, signerRule, isValidRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractAuthValidatedSigner)
				if err := _ContractAuth.contract.UnpackLog(event, "ValidatedSigner", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseValidatedSigner is a log parse operation binding the contract event 0xb0a1cb3375368319eda37d9ea023fe4d12b0185fed7e4022ff1729b4bf40b40b.
//
// Solidity: event ValidatedSigner(address indexed owner, address indexed signer, bytes4 indexed isValid)
func (_ContractAuth *ContractAuthFilterer) ParseValidatedSigner(log types.Log) (*ContractAuthValidatedSigner, error) {
	event := new(ContractAuthValidatedSigner)
	if err := _ContractAuth.contract.UnpackLog(event, "ValidatedSigner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
