package contract

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"math/big"
)

type MultiCallResult struct {
	BlockNumber *big.Int
	ReturnData  []AggregateMulticallResult
}

// CallStaticMulticall is a free data retrieval call binding the contract method 0x1749e1e3.
//
// Solidity: function multicall((address,uint256,bytes)[] calls) returns(uint256 blockNumber, (bool,uint256,bytes)[] returnData)
func (_AggregateMultiCallContract *AggregateMultiCallContractCaller) CallStaticMulticall(
	opts *bind.CallOpts,
	calls []AggregateMulticallCall,
) (result *MultiCallResult, err error) {
	result = new(MultiCallResult)
	var anyList = []any{result}
	err = _AggregateMultiCallContract.contract.Call(opts, &anyList, "multicall", calls)
	if err != nil {
		return nil, err
	}
	if len(anyList) == 0 {
		return nil, nil
	}
	return
}
