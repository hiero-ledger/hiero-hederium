package rpc

import (
	"context"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/service"
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
	m.registerDebugMethods()

	return m
}

func (m *Methods) GetMethod(name string) (MethodInfo, bool) {
	method, ok := m.methods[name]
	return method, ok
}

func (m *Methods) registerMethod(info MethodInfo) {
	m.methods[info.Name] = info
}

func (m *Methods) registerEthMethods() {
	m.registerMethod(MethodInfo{
		Name: "eth_blockNumber",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetBlockNumber()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getBlockByHash",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetBlockByHashParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetBlockByHashParams)
			return services.EthService().GetBlockByHash(p.BlockHash, p.ShowDetails)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getBlockByNumber",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetBlockByNumberParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetBlockByNumberParams)
			return services.EthService().GetBlockByNumber(p.BlockNumber, p.ShowDetails)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getBalance",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetBalanceParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetBalanceParams)
			return services.EthService().GetBalance(p.Address, p.BlockNumber), nil
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getTransactionCount",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetTransactionCountParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetTransactionCountParams)
			return services.EthService().GetTransactionCount(p.Address, p.BlockNumber), nil
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getCode",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetCodeParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetCodeParams)
			return services.EthService().GetCode(p.Address, p.BlockNumber)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getStorageAt",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetStorageAtParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetStorageAtParams)
			return services.EthService().GetStorageAt(p.Address, p.StoragePosition, p.BlockNumber)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_sendRawTransaction",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthSendRawTransactionParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthSendRawTransactionParams)
			return services.EthService().SendRawTransaction(p.SignedTransaction)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getTransactionByHash",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetTransactionByHashParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetTransactionByHashParams)
			return services.EthService().GetTransactionByHash(p.TransactionHash)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getTransactionReceipt",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetTransactionReceiptParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetTransactionReceiptParams)
			return services.EthService().GetTransactionReceipt(p.TransactionHash)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getBlockTransactionCountByHash",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetBlockTransactionCountByHashParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetBlockTransactionCountByHashParams)
			return services.EthService().GetBlockTransactionCountByHash(p.BlockHash)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getBlockTransactionCountByNumber",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetBlockTransactionCountByNumberParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetBlockTransactionCountByNumberParams)
			return services.EthService().GetBlockTransactionCountByNumber(p.BlockNumber)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getTransactionByBlockHashAndIndex",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetTransactionByBlockHashAndIndexParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetTransactionByBlockHashAndIndexParams)
			return services.EthService().GetTransactionByBlockHashAndIndex(p.BlockHash, p.TransactionIndex)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getTransactionByBlockNumberAndIndex",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetTransactionByBlockNumberAndIndexParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetTransactionByBlockNumberAndIndexParams)
			return services.EthService().GetTransactionByBlockNumberAndIndex(p.BlockNumber, p.TransactionIndex)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_call",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthCallParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthCallParams)
			return services.EthService().Call(p.CallObject, p.Block)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_estimateGas",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthEstimateGasParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthEstimateGasParams)
			return services.EthService().EstimateGas(p.CallObject, p.BlockParameter)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_gasPrice",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetGasPrice()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_chainId",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetChainId()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getLogs",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetLogsParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetLogsParams)
			logParams := p.ToLogParams()
			return services.EthService().GetLogs(logParams)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_feeHistory",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthFeeHistoryParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthFeeHistoryParams)
			return services.EthService().FeeHistory(p.BlockCount, p.NewestBlock, p.RewardPercentiles)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getUncleCountByBlockHash",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetUncleCountByBlockHash("")
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getUncleCountByBlockNumber",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetUncleCountByBlockNumber("")
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getUncleByBlockHashAndIndex",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetUncleByBlockHashAndIndex("", "")
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getUncleByBlockNumberAndIndex",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetUncleByBlockNumberAndIndex("", "")
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_accounts",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().GetAccounts()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_syncing",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().Syncing()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_mining",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().Mining()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_maxPriorityFeePerGas",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().MaxPriorityFeePerGas()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_hashrate",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().Hashrate()
		},
	})
}

// registerWeb3Methods registers all Web3 API methods
func (m *Methods) registerWeb3Methods() {
	m.registerMethod(MethodInfo{
		Name: "web3_clientVersion",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.Web3Service().ClientVersion(), nil
		},
	})

	m.registerMethod(MethodInfo{
		Name: "web3_client_version",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.Web3Service().ClientVersion(), nil
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_submitWork",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.EthService().SubmitWork()
		},
	})
}

// registerNetMethods registers all Net API methods
func (m *Methods) registerNetMethods() {
	m.registerMethod(MethodInfo{
		Name: "net_listening",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.NetService().Listening(), nil
		},
	})

	m.registerMethod(MethodInfo{
		Name: "net_version",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.NetService().Version(), nil
		},
	})
}

// registerFilterMethods registers all Filter API methods
func (m *Methods) registerFilterMethods() {
	m.registerMethod(MethodInfo{
		Name: "eth_newFilter",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthNewFilterParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthNewFilterParams)
			return services.FilterService().NewFilter(p.FromBlock, p.ToBlock, p.Address, p.Topics)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_newBlockFilter",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.FilterService().NewBlockFilter()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_uninstallFilter",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthUninstallFilterParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthUninstallFilterParams)
			return services.FilterService().UninstallFilter(p.FilterID)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_newPendingTransactionFilter",
		ParamCreator: func() domain.RPCParams {
			return &domain.NoParameters{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			return services.FilterService().NewPendingTransactionFilter()
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getFilterLogs",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetFilterLogsParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetFilterLogsParams)
			return services.FilterService().GetFilterLogs(p.FilterID)
		},
	})

	m.registerMethod(MethodInfo{
		Name: "eth_getFilterChanges",
		ParamCreator: func() domain.RPCParams {
			return &domain.EthGetFilterChangesParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.EthGetFilterChangesParams)
			return services.FilterService().GetFilterChanges(p.FilterID)
		},
	})
}

func (m *Methods) registerDebugMethods() {
	m.registerMethod(MethodInfo{
		Name: "debug_traceTransaction",
		ParamCreator: func() domain.RPCParams {
			return &domain.DebugTraceTransactionParams{}
		},
		Handler: func(ctx context.Context, params domain.RPCParams, services service.ServiceProvider) (interface{}, *domain.RPCError) {
			p := params.(*domain.DebugTraceTransactionParams)
			return services.DebugService().DebugTraceTransaction(p.TransactionIDOrHash, p.Tracer, p.Config)
		},
	})
}
