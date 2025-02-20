package transport

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"github.com/LimeChain/Hederium/internal/service"
)

var ethService *service.EthService
var web3Service *service.Web3Service
var netService *service.NetService
var logger *zap.Logger
var filterService service.FilterService
var commonService service.CommonService

func SetupRouter(
	hClient *hedera.HederaClient,
	mClient *hedera.MirrorClient,
	log *zap.Logger,
	applicationVersion string,
	chainId string,
	apiKeyStore *limiter.APIKeyStore,
	tieredLimiter *limiter.TieredLimiter,
	enforceAPIKey bool,
	cacheService cache.CacheService,
) *gin.Engine {
	logger = log
	commonService = service.NewCommonService(mClient, log, cacheService)
	ethService = service.NewEthService(hClient, mClient, commonService, log, tieredLimiter, chainId, cacheService)
	web3Service = service.NewWeb3Service(log, applicationVersion)
	netService = service.NewNetService(log, chainId)
	filterService = service.NewFilterService(mClient, cacheService, log, commonService)
	router := gin.Default()

	if enforceAPIKey {
		router.POST("/", AuthAndRateLimitMiddleware(apiKeyStore, tieredLimiter), rpcHandler)
	} else {
		router.POST("/", rpcHandler)
	}

	return router
}
