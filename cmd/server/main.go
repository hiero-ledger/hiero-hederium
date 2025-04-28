package main

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/LimeChain/Hederium/internal/infrastructure/config"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"github.com/LimeChain/Hederium/internal/infrastructure/logger"
	"github.com/LimeChain/Hederium/internal/infrastructure/startup"
	"github.com/LimeChain/Hederium/internal/transport/http_server"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		return
	}
	log := logger.InitLogger(viper.GetString("logging.level"))
	defer log.Sync()

	// Log startup information
	startup.LogStartup()

	hClient, err := hedera.NewHederaClient(
		viper.GetString("hedera.network"),
		viper.GetString("hedera.operatorId"),
		viper.GetString("hedera.operatorKey"),
	)
	if err != nil {
		log.Fatal("Failed to initialize Hedera client", zap.Error(err))
	}

	applicationVersion := viper.GetString("application.version")
	chainId := viper.GetString("hedera.chainId")
	apiKeyStore := limiter.NewAPIKeyStore(viper.Get("apiKeys"))
	tieredLimiter := limiter.NewTieredLimiter(viper.GetStringMap("limiter"), viper.GetInt("hedera.hbarBudget"))

	cacheService := cache.NewMemoryCache(viper.GetDuration("cache.defaultExpiration"), viper.GetDuration("cache.cleanupInterval"))

	mClient := hedera.NewMirrorClient(viper.GetString("mirrorNode.baseUrl"), viper.GetInt("mirrorNode.timeoutSeconds"), log, cacheService)

	enforceAPIKey := viper.GetBool("features.enforceApiKey")
	enableBatchRequests := viper.GetBool("features.enableBatchRequests")

	port := viper.GetString("server.port")

	server := http_server.NewServer(hClient, mClient, log, applicationVersion, chainId, apiKeyStore, tieredLimiter, enforceAPIKey, enableBatchRequests, cacheService, port)
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
