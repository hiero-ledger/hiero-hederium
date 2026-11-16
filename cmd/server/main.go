package main

import (
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/cache"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/config"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/hedera"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/limiter"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/logger"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/startup"
	"github.com/hiero-ledger/hiero-hederium/internal/transport/http_server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	log := logger.InitLogger(cfg.Logging.Level)
	defer func() { _ = log.Sync() }()

	// Log startup information
	startup.LogStartup(cfg)

	hClient, err := hedera.NewHederaClient(
		cfg.Hedera.Network,
		cfg.Hedera.OperatorID,
		cfg.Hedera.OperatorKey.Reveal(),
	)
	if err != nil {
		// log.Fatal exits immediately; ensure sync happens before exiting
		log.Error("Failed to initialize Hedera client", zap.Error(err))
		return
	}

	apiKeyStore := limiter.NewAPIKeyStore(cfg.APIKeys)
	tieredLimiter := limiter.NewTieredLimiter(cfg.Limiter, cfg.Hedera.HbarBudget)

	cacheService := cache.NewMemoryCache(cfg.Cache.DefaultExpiration, cfg.Cache.CleanupInterval)

	mClient := hedera.NewMirrorClientWithTimeout(cfg.MirrorNode.BaseURL, cfg.MirrorNode.Timeout, log, cacheService)

	server := http_server.NewServer(
		hClient, mClient, log,
		cfg.Application.Version, cfg.Hedera.ChainID,
		apiKeyStore, tieredLimiter,
		cfg.Features.EnforceAPIKey, cfg.Features.EnableBatchRequests,
		cacheService, strconv.Itoa(cfg.Server.Port),
	)
	if err := server.Start(); err != nil {
		log.Error("Failed to start server", zap.Error(err))
		return
	}
}
