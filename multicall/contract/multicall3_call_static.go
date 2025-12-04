package contract

import "github.com/ethereum/go-ethereum/accounts/abi/bind"

func (mc *Multicall3Caller) Aggregate3Static(
	opts *bind.CallOpts,
	calls []Multicall3Call3,
) (result *MultiCall3Result, err error) {
	result = new(MultiCall3Result)
	var anyList = []any{result}
	err = mc.contract.Call(opts, &anyList, "aggregate3", calls)
	if err != nil {
		return nil, err
	}
	if len(anyList) == 0 {
		return nil, nil
	}
	return
}

type MultiCall3Result struct {
	ReturnData []Multicall3Result
}
