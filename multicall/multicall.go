package multicall

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hawkneo/utils/multicall/contract"
)

type Multicall struct {
	eth *contract.AggregateMultiCallContractCaller
}

func NewMulticall(eth bind.ContractCaller, addr common.Address) (*Multicall, error) {
	caller, err := contract.NewAggregateMultiCallContractCaller(addr, eth)
	if err != nil {
		return nil, err
	}
	return &Multicall{
		eth: caller,
	}, nil
}

type CallResult struct {
	Call    ViewCall
	Raw     []byte
	Decoded []interface{}
	Error   error
}

type Result struct {
	BlockNumber uint64
	Calls       []CallResult
}

func (mc *Multicall) Call(opts *bind.CallOpts, calls ViewCalls) (*Result, error) {
	resultRaw, err := mc.makeRequest(opts, calls)
	if err != nil {
		return nil, err
	}
	return calls.decode(resultRaw)
}

func (mc *Multicall) makeRequest(opts *bind.CallOpts, calls ViewCalls) (*contract.MultiCallResult, error) {
	callDatas := make([]contract.AggregateMulticallCall, 0)
	for _, call := range calls {
		data, err := call.CallData()
		if err != nil {
			return nil, err
		}
		callDatas = append(callDatas, contract.AggregateMulticallCall{
			Target:   call.Target(),
			GasLimit: call.GasLimit(),
			CallData: data,
		})
	}
	return mc.eth.CallStaticMulticall(opts, callDatas)
}
