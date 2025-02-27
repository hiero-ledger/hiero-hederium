package service

import (
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"go.uber.org/zap"
)

type ServiceProvider interface {
	EthService() *EthService
	Web3Service() Web3Servicer
	NetService() NetServicer
	FilterService() FilterServicer
}

// For now we use *EthService instead of EthServicer
type serviceProvider struct {
	ethService    *EthService
	web3Service   Web3Servicer
	netService    NetServicer
	filterService FilterServicer
}

func NewServiceProvider(
	hClient *hedera.HederaClient,
	mClient *hedera.MirrorClient,
	log *zap.Logger,
	applicationVersion string,
	chainId string,
	apiKeyStore *limiter.APIKeyStore,
	tieredLimiter *limiter.TieredLimiter,
	cacheService cache.CacheService,
) ServiceProvider {
	commonService := NewCommonService(mClient, log, cacheService)
	ethService := NewEthService(hClient, mClient, commonService, log, tieredLimiter, chainId, cacheService)
	web3Service := NewWeb3Service(log, applicationVersion)
	netService := NewNetService(log, chainId)
	filterService := NewFilterService(mClient, cacheService, log, commonService)

	return &serviceProvider{ethService: ethService, web3Service: web3Service, netService: netService, filterService: filterService}
}

func (s *serviceProvider) EthService() *EthService {
	return s.ethService
}

func (s *serviceProvider) Web3Service() Web3Servicer {
	return s.web3Service
}

func (s *serviceProvider) NetService() NetServicer {
	return s.netService
}

func (s *serviceProvider) FilterService() FilterServicer {
	return s.filterService
}
