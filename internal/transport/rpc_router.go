package transport

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/georgi-l95/Hederium/internal/infrastructure/hedera"
	"github.com/georgi-l95/Hederium/internal/infrastructure/limiter"
	"github.com/georgi-l95/Hederium/internal/service"
	sdkhedera "github.com/hashgraph/hedera-sdk-go/v2"
)

var ethService *service.EthService
var logger *zap.Logger

func SetupRouter(
	hClient *sdkhedera.Client,
	mClient *hedera.MirrorClient,
	log *zap.Logger,
	apiKeyStore *limiter.APIKeyStore,
	tieredLimiter *limiter.TieredLimiter,
	enforceAPIKey bool,
) *gin.Engine {
	logger = log
	ethService = service.NewEthService(hClient, mClient, log, tieredLimiter)

	router := gin.Default()

	if enforceAPIKey {
		router.POST("/", AuthAndRateLimitMiddleware(apiKeyStore, tieredLimiter), rpcHandler)
	} else {
		router.POST("/", rpcHandler)
	}

	return router
}
