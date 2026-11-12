package rpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hiero-ledger/hiero-hederium/internal/domain"
	"github.com/hiero-ledger/hiero-hederium/internal/transport/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registeredMethods is every method Hederium serves, with the params type its
// handler decodes into.
var registeredMethods = map[string]string{
	"eth_blockNumber":                         "*domain.NoParameters",
	"eth_getBlockByHash":                      "*domain.EthGetBlockByHashParams",
	"eth_getBlockByNumber":                    "*domain.EthGetBlockByNumberParams",
	"eth_getBalance":                          "*domain.EthGetBalanceParams",
	"eth_getTransactionCount":                 "*domain.EthGetTransactionCountParams",
	"eth_getCode":                             "*domain.EthGetCodeParams",
	"eth_getStorageAt":                        "*domain.EthGetStorageAtParams",
	"eth_sendRawTransaction":                  "*domain.EthSendRawTransactionParams",
	"eth_getTransactionByHash":                "*domain.EthGetTransactionByHashParams",
	"eth_getTransactionReceipt":               "*domain.EthGetTransactionReceiptParams",
	"eth_getBlockTransactionCountByHash":      "*domain.EthGetBlockTransactionCountByHashParams",
	"eth_getBlockTransactionCountByNumber":    "*domain.EthGetBlockTransactionCountByNumberParams",
	"eth_getTransactionByBlockHashAndIndex":   "*domain.EthGetTransactionByBlockHashAndIndexParams",
	"eth_getTransactionByBlockNumberAndIndex": "*domain.EthGetTransactionByBlockNumberAndIndexParams",
	"eth_call":                                "*domain.EthCallParams",
	"eth_estimateGas":                         "*domain.EthEstimateGasParams",
	"eth_gasPrice":                            "*domain.NoParameters",
	"eth_chainId":                             "*domain.NoParameters",
	"eth_getLogs":                             "*domain.EthGetLogsParams",
	"eth_feeHistory":                          "*domain.EthFeeHistoryParams",
	"eth_getUncleCountByBlockHash":            "*domain.EthGetUncleCountByBlockHashParams",
	"eth_getUncleCountByBlockNumber":          "*domain.EthGetUncleCountByBlockNumberParams",
	"eth_getUncleByBlockHashAndIndex":         "*domain.EthGetUncleByBlockHashAndIndexParams",
	"eth_getUncleByBlockNumberAndIndex":       "*domain.EthGetUncleByBlockNumberAndIndexParams",
	"eth_accounts":                            "*domain.NoParameters",
	"eth_syncing":                             "*domain.NoParameters",
	"eth_mining":                              "*domain.NoParameters",
	"eth_maxPriorityFeePerGas":                "*domain.NoParameters",
	"eth_hashrate":                            "*domain.NoParameters",
	"web3_clientVersion":                      "*domain.NoParameters",
	"net_listening":                           "*domain.NoParameters",
	"net_version":                             "*domain.NoParameters",
	"eth_newFilter":                           "*domain.EthNewFilterParams",
	"eth_newBlockFilter":                      "*domain.NoParameters",
	"eth_uninstallFilter":                     "*domain.EthUninstallFilterParams",
	"eth_newPendingTransactionFilter":         "*domain.NoParameters",
	"eth_getFilterLogs":                       "*domain.EthGetFilterLogsParams",
	"eth_getFilterChanges":                    "*domain.EthGetFilterChangesParams",
}

func TestMethods_RegistersEveryMethodWithItsParamsType(t *testing.T) {
	methods := rpc.NewMethods()

	for name, paramsType := range registeredMethods {
		info, ok := methods.GetMethod(name)
		require.True(t, ok, "%s is not registered", name)
		assert.Equal(t, name, info.Name)
		assert.Equal(t, paramsType, fmt.Sprintf("%T", info.ParamCreator()), name)
	}

	_, ok := methods.GetMethod("eth_unknownMethod")
	assert.False(t, ok)
}

// A handler given params of another type answers with an internal error
// instead of panicking on the type assertion.
func TestMethods_HandlerRejectsParamsOfAnotherType(t *testing.T) {
	methods := rpc.NewMethods()

	for name, paramsType := range registeredMethods {
		if paramsType == "*domain.NoParameters" {
			continue // these handlers ignore their params
		}
		info, ok := methods.GetMethod(name)
		require.True(t, ok, name)

		var result interface{}
		var errRpc *domain.RPCError
		require.NotPanics(t, func() {
			result, errRpc = info.Handler(context.Background(), &domain.NoParameters{}, fakeServiceProvider{})
		}, name)
		assert.Nil(t, result, name)
		require.NotNil(t, errRpc, name)
		assert.Equal(t, domain.InternalError, errRpc.Code, name)
	}
}
