package multicall

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ABIViewCall struct {
	method       abi.Method
	target       common.Address
	gasLimit     *big.Int // used in multi call
	allowFailure bool     // used in multi call3
	arguments    []interface{}
	callback     func(err error, returnValues []interface{}) error // callback for each call
}

func (A *ABIViewCall) Callback() func(err error, returnValues []interface{}) error {
	return A.callback
}

func (A *ABIViewCall) GasLimit() *big.Int {
	return A.gasLimit
}

func (A *ABIViewCall) Target() common.Address {
	return A.target
}

func (A *ABIViewCall) AllowFailure() bool {
	return A.allowFailure
}

func (A *ABIViewCall) CallData() ([]byte, error) {
	callData := make([]byte, 0)
	callData = append(callData, A.method.ID...)
	argsCallData, err := A.method.Inputs.Pack(A.arguments...)
	if err != nil {
		return nil, err
	}
	callData = append(callData, argsCallData...)
	return callData, nil
}

func (A *ABIViewCall) Decode(raw []byte) ([]interface{}, error) {
	return A.method.Outputs.Unpack(raw)
}

func NewABIViewCall(target common.Address, method abi.Method, arguments []interface{}, callback func(err error, returnValues []interface{}) error) ViewCall {
	return &ABIViewCall{
		method:    method,
		target:    target,
		arguments: arguments,
		callback:  callback,
	}
}

func NewABIViewCallWithAllowFailure(target common.Address, method abi.Method, arguments []interface{}, callback func(err error, returnValues []interface{}) error, allowFailure bool) ViewCall {
	return &ABIViewCall{
		target:       target,
		method:       method,
		arguments:    arguments,
		allowFailure: allowFailure,
		callback:     callback,
	}
}

func NewABIViewCallWithGasLimit(target common.Address, method abi.Method, arguments []interface{}, callback func(err error, returnValues []interface{}) error, gasLimit *big.Int) ViewCall {
	return &ABIViewCall{
		target:    target,
		method:    method,
		arguments: arguments,
		gasLimit:  gasLimit,
		callback:  callback,
	}
}
