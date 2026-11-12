package rpc

import (
	"context"

	"github.com/hiero-ledger/hiero-hederium/internal/domain"
	"github.com/hiero-ledger/hiero-hederium/internal/service"
)

type HandlerFunc func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError)

type MethodInfo struct {
	Name         string
	ParamCreator func() domain.RPCParams
	Handler      HandlerFunc
}

type Methods struct {
	methods map[string]MethodInfo
}

func NewMethods() *Methods {
	m := &Methods{
		methods: make(map[string]MethodInfo),
	}
	m.registerEthMethods()
	m.registerWeb3Methods()
	m.registerNetMethods()
	m.registerFilterMethods()

	return m
}

func (m *Methods) GetMethod(name string) (MethodInfo, bool) {
	method, ok := m.methods[name]
	return method, ok
}

func (m *Methods) registerMethod(info MethodInfo) {
	m.methods[info.Name] = info
}

// method builds a MethodInfo whose ParamCreator and Handler share the params
// type P, so the handler receives exactly the type ParamCreator made. Params of
// any other type are answered with an internal error instead of a panic.
func method[T any, P interface {
	*T
	domain.RPCParams
}](name string, handle func(ctx context.Context, p P, services service.ServiceProvider) (interface{}, *domain.RPCError)) MethodInfo {
	return MethodInfo{
		Name: name,
		ParamCreator: func() domain.RPCParams {
			return P(new(T))
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p, ok := params.(P)
			if !ok {
				return nil, domain.NewInternalError("Internal error")
			}
			return handle(ctx, p, services)
		},
	}
}

func (m *Methods) registerEthMethods() {
	m.registerMethod(method("eth_blockNumber", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetBlockNumber()
	}))

	m.registerMethod(method("eth_getBlockByHash", func(ctx context.Context, p *domain.EthGetBlockByHashParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetBlockByHash(p.BlockHash, p.ShowDetails)
	}))

	m.registerMethod(method("eth_getBlockByNumber", func(ctx context.Context, p *domain.EthGetBlockByNumberParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetBlockByNumber(p.BlockNumber, p.ShowDetails)
	}))

	m.registerMethod(method("eth_getBalance", func(ctx context.Context, p *domain.EthGetBalanceParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetBalance(p.Address, p.BlockNumber)
	}))

	m.registerMethod(method("eth_getTransactionCount", func(ctx context.Context, p *domain.EthGetTransactionCountParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetTransactionCount(p.Address, p.BlockNumber), nil
	}))

	m.registerMethod(method("eth_getCode", func(ctx context.Context, p *domain.EthGetCodeParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetCode(p.Address, p.BlockNumber)
	}))

	m.registerMethod(method("eth_getStorageAt", func(ctx context.Context, p *domain.EthGetStorageAtParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetStorageAt(p.Address, p.StoragePosition, p.BlockNumber)
	}))

	m.registerMethod(method("eth_sendRawTransaction", func(ctx context.Context, p *domain.EthSendRawTransactionParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().SendRawTransaction(p.SignedTransaction)
	}))

	m.registerMethod(method("eth_getTransactionByHash", func(ctx context.Context, p *domain.EthGetTransactionByHashParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetTransactionByHash(p.TransactionHash)
	}))

	m.registerMethod(method("eth_getTransactionReceipt", func(ctx context.Context, p *domain.EthGetTransactionReceiptParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetTransactionReceipt(p.TransactionHash)
	}))

	m.registerMethod(method("eth_getBlockTransactionCountByHash", func(ctx context.Context, p *domain.EthGetBlockTransactionCountByHashParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetBlockTransactionCountByHash(p.BlockHash)
	}))

	m.registerMethod(method("eth_getBlockTransactionCountByNumber", func(ctx context.Context, p *domain.EthGetBlockTransactionCountByNumberParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetBlockTransactionCountByNumber(p.BlockNumber)
	}))

	m.registerMethod(method("eth_getTransactionByBlockHashAndIndex", func(ctx context.Context, p *domain.EthGetTransactionByBlockHashAndIndexParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetTransactionByBlockHashAndIndex(p.BlockHash, p.TransactionIndex)
	}))

	m.registerMethod(method("eth_getTransactionByBlockNumberAndIndex", func(ctx context.Context, p *domain.EthGetTransactionByBlockNumberAndIndexParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetTransactionByBlockNumberAndIndex(p.BlockNumber, p.TransactionIndex)
	}))

	m.registerMethod(method("eth_call", func(ctx context.Context, p *domain.EthCallParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().Call(p.CallObject, p.Block)
	}))

	m.registerMethod(method("eth_estimateGas", func(ctx context.Context, p *domain.EthEstimateGasParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().EstimateGas(p.CallObject, p.BlockParameter)
	}))

	m.registerMethod(method("eth_gasPrice", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetGasPrice()
	}))

	m.registerMethod(method("eth_chainId", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetChainId()
	}))

	m.registerMethod(method("eth_getLogs", func(ctx context.Context, p *domain.EthGetLogsParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		logParams := p.ToLogParams()
		return services.EthService().GetLogs(logParams)
	}))

	m.registerMethod(method("eth_feeHistory", func(ctx context.Context, p *domain.EthFeeHistoryParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().FeeHistory(p.BlockCount, p.NewestBlock, p.RewardPercentiles)
	}))

	m.registerMethod(method("eth_getUncleCountByBlockHash", func(ctx context.Context, p *domain.EthGetUncleCountByBlockHashParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetUncleCountByBlockHash(p.BlockHash)
	}))

	m.registerMethod(method("eth_getUncleCountByBlockNumber", func(ctx context.Context, p *domain.EthGetUncleCountByBlockNumberParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetUncleCountByBlockNumber(p.BlockNumber)
	}))

	m.registerMethod(method("eth_getUncleByBlockHashAndIndex", func(ctx context.Context, p *domain.EthGetUncleByBlockHashAndIndexParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetUncleByBlockHashAndIndex(p.BlockHash, p.Index)
	}))

	m.registerMethod(method("eth_getUncleByBlockNumberAndIndex", func(ctx context.Context, p *domain.EthGetUncleByBlockNumberAndIndexParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetUncleByBlockNumberAndIndex(p.BlockNumber, p.Index)
	}))

	m.registerMethod(method("eth_accounts", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().GetAccounts()
	}))

	m.registerMethod(method("eth_syncing", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().Syncing()
	}))

	m.registerMethod(method("eth_mining", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().Mining()
	}))

	m.registerMethod(method("eth_maxPriorityFeePerGas", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().MaxPriorityFeePerGas()
	}))

	m.registerMethod(method("eth_hashrate", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.EthService().Hashrate()
	}))
}

// registerWeb3Methods registers all Web3 API methods
func (m *Methods) registerWeb3Methods() {
	m.registerMethod(method("web3_clientVersion", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.Web3Service().ClientVersion(), nil
	}))
}

// registerNetMethods registers all Net API methods
func (m *Methods) registerNetMethods() {
	m.registerMethod(method("net_listening", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.NetService().Listening(), nil
	}))

	m.registerMethod(method("net_version", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.NetService().Version(), nil
	}))
}

// registerFilterMethods registers all Filter API methods
func (m *Methods) registerFilterMethods() {
	m.registerMethod(method("eth_newFilter", func(ctx context.Context, p *domain.EthNewFilterParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.FilterService().NewFilter(p.FromBlock, p.ToBlock, p.Address, p.Topics)
	}))

	m.registerMethod(method("eth_newBlockFilter", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.FilterService().NewBlockFilter()
	}))

	m.registerMethod(method("eth_uninstallFilter", func(ctx context.Context, p *domain.EthUninstallFilterParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.FilterService().UninstallFilter(p.FilterID)
	}))

	m.registerMethod(method("eth_newPendingTransactionFilter", func(ctx context.Context, _ *domain.NoParameters, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.FilterService().NewPendingTransactionFilter()
	}))

	m.registerMethod(method("eth_getFilterLogs", func(ctx context.Context, p *domain.EthGetFilterLogsParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.FilterService().GetFilterLogs(p.FilterID)
	}))

	m.registerMethod(method("eth_getFilterChanges", func(ctx context.Context, p *domain.EthGetFilterChangesParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
		return services.FilterService().GetFilterChanges(p.FilterID)
	}))
}
