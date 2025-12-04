package multicall

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"math/big"
	"strings"
	"testing"
)

var (
	ABI = "[{\"inputs\":[{\"internalType\":\"contractIPositionRouterState\",\"name\":\"_positionRouter\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"_max\",\"type\":\"uint128\"}],\"name\":\"calculateNextMulticall\",\"outputs\":[{\"internalType\":\"contractIPool[]\",\"name\":\"pools\",\"type\":\"address[]\"},{\"components\":[{\"internalType\":\"uint128\",\"name\":\"index\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"indexNext\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"indexEnd\",\"type\":\"uint128\"}],\"internalType\":\"structExecutorAssistant.IndexPerOperation[7]\",\"name\":\"indexPerOperations\",\"type\":\"tuple[7]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"positionRouter\",\"outputs\":[{\"internalType\":\"contractIPositionRouterState\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
)

func TestMulticall(t *testing.T) {
	mainNet := "https://eth.llamarpc.com"
	client, _ := ethclient.Dial(mainNet)
	muticallAddress := common.HexToAddress("0xdb48358424d147804631092a5F568Fc2332d7248")
	contract, err := NewMulticall(client, muticallAddress)
	require.NoError(t, err)
	res, err := contract.Call(nil, ViewCalls{
		NewViewCallWithGasLimit(common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"), "name()(string)", []interface{}{}, func(err error, returnValues []interface{}) error {
			require.NoError(t, err)
			require.Len(t, returnValues, 1)
			require.IsType(t, "", returnValues[0])
			require.Equal(t, "Tether USD", returnValues[0].(string))
			return nil
		}, big.NewInt(100000)),
		NewViewCallWithGasLimit(common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"), "balanceOf(address)(uint256)", []interface{}{common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")}, func(err error, returnValues []interface{}) error {
			require.NoError(t, err)
			require.Len(t, returnValues, 1)
			require.IsType(t, big.NewInt(0), returnValues[0])
			require.NotZero(t, returnValues[0].(*big.Int).Uint64())
			return nil
		}, big.NewInt(100000)),
		NewViewCallWithGasLimit(common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"), "balanceOf(address)(uint256)", []interface{}{common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")}, func(err error, returnValues []interface{}) error {
			require.Error(t, err)
			return nil
		}, big.NewInt(100)),
	})
	require.NoError(t, err)
	for _, result := range res.Calls {
		require.NoError(t, result.Error)
	}

}

func TestMulticall3(t *testing.T) {
	mainNet := "https://eth.llamarpc.com"
	client, _ := ethclient.Dial(mainNet)
	muticall3Address := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
	contract, err := NewMulticall3(client, muticall3Address)
	require.NoError(t, err)
	res, err := contract.Call(nil, ViewCalls{
		NewViewCall(common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"), "name()(string)", []interface{}{}, func(err error, returnValues []interface{}) error {
			require.NoError(t, err)
			require.Len(t, returnValues, 1)
			require.IsType(t, "", returnValues[0])
			require.Equal(t, "Tether USD", returnValues[0].(string))
			return nil
		}),
		NewViewCall(
			common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
			"balanceOf(address)(uint256)",
			[]interface{}{common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")},
			func(err error, returnValues []interface{}) error {
				require.NoError(t, err)
				require.Len(t, returnValues, 1)
				require.IsType(t, big.NewInt(0), returnValues[0])
				require.NotZero(t, returnValues[0].(*big.Int).Uint64())
				return nil
			},
		),
		NewViewCallWithDecoder(
			common.HexToAddress("0xd46eC47D776f47a9A2dbBA3F36F38e0df521a19a"),
			"calculateNextMulticall(uint128)(address[],IndexPerOperation[7])",
			[]interface{}{big.NewInt(10)},
			func(raw []byte) (returnValues []interface{}, err error) {
				return nil, nil
			},
			func(err error, returnValues []interface{}) error {
				return nil
			},
		),
	})
	require.NoError(t, err)
	for _, result := range res.Calls {
		require.NoError(t, result.Error)
	}
}

func TestMulticall3_Decoder(t *testing.T) {
	client, err := ethclient.Dial("https://arbitrum-goerli.publicnode.com")
	require.NoError(t, err)
	contract, err := NewMulticall3(client, common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"))
	require.NoError(t, err)

	parsed, err := abi.JSON(strings.NewReader(ABI))
	require.NoError(t, err)

	_ = parsed
	res, err := contract.Call(nil, ViewCalls{
		NewViewCall(
			common.HexToAddress("0x68bA976974Bb7aed3a5eAa4EB1501dda5D55fBba"),
			"name()(string)",
			[]interface{}{},
			func(err error, returnValues []interface{}) error {
				require.NoError(t, err)
				require.Len(t, returnValues, 1)
				require.IsType(t, "", returnValues[0])
				require.Equal(t, "Equation", returnValues[0].(string))
				return nil
			},
		),
		NewViewCallWithDecoder(
			common.HexToAddress("0xd46eC47D776f47a9A2dbBA3F36F38e0df521a19a"),
			"calculateNextMulticall(uint128)(address[],IndexPerOperation[7])",
			[]interface{}{big.NewInt(10)},
			func(raw []byte) (returnValues []interface{}, err error) {
				out, err := parsed.Unpack("calculateNextMulticall", raw)
				if err != nil {
					return nil, err
				}
				require.Equal(t, 70, len(out[0].([]common.Address)))
				require.Equal(t, 7, len(out[1].([7]interface{})))
				return nil, nil
			},
			func(err error, returnValues []interface{}) error {
				require.NoError(t, err)
				return nil
			},
		),
		NewABIViewCall(common.HexToAddress("0xd46eC47D776f47a9A2dbBA3F36F38e0df521a19a"),
			parsed.Methods["calculateNextMulticall"],
			[]interface{}{big.NewInt(10)},
			func(err error, returnValues []interface{}) error {
				require.NoError(t, err)
				require.Equal(t, 70, len(returnValues[0].([]common.Address)))
				require.Equal(t, 7, len(returnValues[1].([7]interface{})))
				return nil
			}),
	})
	require.NoError(t, err)
	for _, result := range res.Calls {
		require.NoError(t, result.Error)
	}
}
