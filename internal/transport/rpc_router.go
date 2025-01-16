package transport

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"github.com/LimeChain/Hederium/internal/service"
)

var ethService *service.EthService
var web3Service *service.Web3Service
var netService *service.NetService
var logger *zap.Logger

func SetupRouter(
	hClient *hedera.HederaClient,
	mClient *hedera.MirrorClient,
	log *zap.Logger,
	applicationVersion string,
	chainId string,
	apiKeyStore *limiter.APIKeyStore,
	tieredLimiter *limiter.TieredLimiter,
	enforceAPIKey bool,
) *gin.Engine {
	logger = log
	ethService = service.NewEthService(hClient, mClient, log, tieredLimiter, chainId)
	web3Service = service.NewWeb3Service(log, applicationVersion)
	netService = service.NewNetService(log, chainId)

	router := gin.Default()

	if enforceAPIKey {
		router.POST("/", AuthAndRateLimitMiddleware(apiKeyStore, tieredLimiter), rpcHandler)
	} else {
		router.POST("/", rpcHandler)
	}

	return router
}
