package multicall

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gridexswap/utils/multicall/contract"
)

type Multicall3 struct {
	eth *contract.Multicall3Caller
}

func NewMulticall3(eth bind.ContractCaller, addr common.Address) (*Multicall3, error) {
	caller, err := contract.NewMulticall3Caller(addr, eth)
	if err != nil {
		return nil, err
	}
	return &Multicall3{
		eth: caller,
	}, nil
}

func (mc *Multicall3) Call(opts *bind.CallOpts, calls ViewCalls) (*Result, error) {
	resultRaw, err := mc.makeRequest(opts, calls)
	if err != nil {
		return nil, err
	}
	return calls.decode3(resultRaw)
}

func (mc *Multicall3) makeRequest(opts *bind.CallOpts, calls ViewCalls) (*contract.MultiCall3Result, error) {
	callDatas := make([]contract.Multicall3Call3, 0)
	for _, call := range calls {
		data, err := call.CallData()
		if err != nil {
			return nil, err
		}
		callDatas = append(callDatas, contract.Multicall3Call3{
			Target:       call.Target(),
			AllowFailure: call.AllowFailure(),
			CallData:     data,
		})
	}
	return mc.eth.Aggregate3Static(opts, callDatas)
}
