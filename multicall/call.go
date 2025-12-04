package multicall

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gridexswap/utils/multicall/contract"
	"math/big"
	"regexp"
	"strings"
)

type ViewCall interface {
	CallData() ([]byte, error)
	Decode(raw []byte) ([]interface{}, error)
	Callback() func(err error, returnValues []interface{}) error
	GasLimit() *big.Int
	Target() common.Address
	AllowFailure() bool
}

type SignatureViewCall struct {
	target       common.Address
	gasLimit     *big.Int // used in multi call
	allowFailure bool     // used in multi call3
	method       string
	arguments    []interface{}
	decoder      func(raw []byte) (returnValues []interface{}, err error) // custom decoder for struct return type, if nil, ViewCall.decode will be used
	callback     func(err error, returnValues []interface{}) error        // callback for each call
}

type ViewCalls []ViewCall

func NewViewCall(target common.Address, method string, arguments []interface{}, callback func(err error, returnValues []interface{}) error) ViewCall {
	return &SignatureViewCall{
		target:    target,
		method:    method,
		arguments: arguments,
		callback:  callback,
	}
}

func NewViewCallWithDecoder(
	target common.Address,
	method string,
	arguments []interface{},
	decoder func(raw []byte) (returnValues []interface{}, err error),
	callback func(err error, returnValues []interface{}) error,
) ViewCall {
	return &SignatureViewCall{
		target:    target,
		method:    method,
		arguments: arguments,
		decoder:   decoder,
		callback:  callback,
	}
}

func NewViewCallWithAllowFailure(target common.Address, method string, arguments []interface{}, callback func(err error, returnValues []interface{}) error, allowFailure bool) ViewCall {
	return &SignatureViewCall{
		target:       target,
		method:       method,
		arguments:    arguments,
		allowFailure: allowFailure,
		callback:     callback,
	}
}

func NewViewCallWithGasLimit(target common.Address, method string, arguments []interface{}, callback func(err error, returnValues []interface{}) error, gasLimit *big.Int) ViewCall {
	return &SignatureViewCall{
		target:    target,
		method:    method,
		arguments: arguments,
		gasLimit:  gasLimit,
		callback:  callback,
	}
}

func (call *SignatureViewCall) Validate() error {
	if _, err := call.argsCallData(); err != nil {
		return err
	}
	return nil
}

func (call *SignatureViewCall) String() string {
	return fmt.Sprintf("%s:%s", call.target, call.method)
}

var insideParens = regexp.MustCompile("\\(.*?\\)")

func (call *SignatureViewCall) argumentTypes() []string {
	rawArgs := insideParens.FindAllString(call.method, -1)[0]
	rawArgs = strings.Replace(rawArgs, "(", "", -1)
	rawArgs = strings.Replace(rawArgs, ")", "", -1)
	if rawArgs == "" {
		return []string{}
	}
	args := strings.Split(rawArgs, ",")
	for index, arg := range args {
		args[index] = strings.Trim(arg, " ")
	}
	return args
}

func (call *SignatureViewCall) returnTypes() []string {
	rawArgs := insideParens.FindAllString(call.method, -1)[1]
	rawArgs = strings.Replace(rawArgs, "(", "", -1)
	rawArgs = strings.Replace(rawArgs, ")", "", -1)
	args := strings.Split(rawArgs, ",")
	for index, arg := range args {
		args[index] = strings.Trim(arg, " ")
	}
	return args
}

func (call *SignatureViewCall) CallData() ([]byte, error) {
	argsSuffix, err := call.argsCallData()
	if err != nil {
		return nil, err
	}
	methodPrefix, err := call.methodCallData()
	if err != nil {
		return nil, err
	}

	payload := make([]byte, 0)
	payload = append(payload, methodPrefix...)
	payload = append(payload, argsSuffix...)

	return payload, nil
}

func (call *SignatureViewCall) methodCallData() ([]byte, error) {
	methodParts := strings.Split(call.method, ")(")
	var method string
	if len(methodParts) > 1 {
		method = fmt.Sprintf("%s)", methodParts[0])
	} else {
		method = methodParts[0]
	}
	hash := crypto.Keccak256([]byte(method))
	return hash[0:4], nil
}

func (call *SignatureViewCall) argsCallData() ([]byte, error) {
	argTypes := call.argumentTypes()
	if len(argTypes) != len(call.arguments) {
		return nil, fmt.Errorf("number of argument types doesn't match with number of arguments with method %s", call.method)
	}
	argumentValues := make([]interface{}, len(call.arguments))
	arguments := make(abi.Arguments, len(call.arguments))

	for index, argTypeStr := range argTypes {
		argType, err := abi.NewType(argTypeStr, "", nil)
		if err != nil {
			return nil, err
		}

		arguments[index] = abi.Argument{Type: argType}
		argumentValues[index] = call.arguments[index]
		if err != nil {
			return nil, err
		}
	}

	return arguments.Pack(argumentValues...)
}

func (call *SignatureViewCall) Decode(raw []byte) ([]interface{}, error) {
	if call.decoder != nil {
		return call.decoder(raw)
	}
	retTypes := call.returnTypes()
	args := make(abi.Arguments, 0, 0)
	for index, retTypeStr := range retTypes {
		retType, err := abi.NewType(retTypeStr, "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, abi.Argument{Name: fmt.Sprintf("ret%d", index), Type: retType})
	}
	decoded := make(map[string]interface{})
	err := args.UnpackIntoMap(decoded, raw)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, len(retTypes))
	for index := range retTypes {
		key := fmt.Sprintf("ret%d", index)
		returns[index] = decoded[key]
	}
	return returns, nil
}

func (call *SignatureViewCall) GasLimit() *big.Int {
	if call.gasLimit == nil {
		return big.NewInt(0)
	}
	return call.gasLimit
}

func (call *SignatureViewCall) Callback() func(err error, returnValues []interface{}) error {
	return call.callback
}

func (call *SignatureViewCall) Target() common.Address {
	return call.target
}

func (call *SignatureViewCall) AllowFailure() bool {
	return call.allowFailure
}

func (calls ViewCalls) decode(callResponse *contract.MultiCallResult) (*Result, error) {
	result := &Result{}
	result.BlockNumber = callResponse.BlockNumber.Uint64()
	result.Calls = make([]CallResult, len(calls))
	for index, call := range calls {
		callResult := CallResult{
			Call: call,
			Raw:  callResponse.ReturnData[index].ReturnData,
		}
		var err error = nil
		var returnValues []interface{} = nil
		if callResponse.ReturnData[index].Success {
			returnValues, err = call.Decode(callResponse.ReturnData[index].ReturnData)
		} else {
			err = fmt.Errorf("call contract error, returns: %s, gaslimit: %s, gasused: %s",
				string(callResponse.ReturnData[index].ReturnData), call.GasLimit(), callResponse.ReturnData[index].GasUsed)
		}
		callResult.Error = err
		callResult.Decoded = returnValues
		if call.Callback() != nil {
			callResult.Error = call.Callback()(err, returnValues)
		}
		result.Calls[index] = callResult
	}
	return result, nil
}

func (calls ViewCalls) decode3(callResponse *contract.MultiCall3Result) (*Result, error) {
	result := &Result{}
	result.BlockNumber = 0
	result.Calls = make([]CallResult, len(calls))
	for index, call := range calls {
		callResult := CallResult{
			Call: call,
			Raw:  callResponse.ReturnData[index].ReturnData,
		}

		var err error = nil
		var returnValues []interface{} = nil
		if callResponse.ReturnData[index].Success {
			returnValues, err = call.Decode(callResponse.ReturnData[index].ReturnData)
		} else {
			err = fmt.Errorf("call contract error, returns: %s",
				string(callResponse.ReturnData[index].ReturnData))
		}
		callResult.Error = err
		callResult.Decoded = returnValues
		if call.Callback() != nil {
			callResult.Error = call.Callback()(err, returnValues)
		}

		result.Calls[index] = callResult
	}
	return result, nil
}
