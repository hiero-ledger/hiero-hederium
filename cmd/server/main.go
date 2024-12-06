package main

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/georgi-l95/Hederium/internal/infrastructure/hedera"
	"github.com/georgi-l95/Hederium/internal/infrastructure/limiter"
	"github.com/georgi-l95/Hederium/internal/infrastructure/logger"
	"github.com/georgi-l95/Hederium/internal/transport"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./configs")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
	}
}

func main() {
	initConfig()
	log := logger.InitLogger(viper.GetString("logging.level"))
	defer log.Sync()

	hClient, err := hedera.NewHederaClient(
		viper.GetString("hedera.network"),
		viper.GetString("hedera.operatorId"),
		viper.GetString("hedera.operatorKey"),
	)
	if err != nil {
		log.Fatal("Failed to initialize Hedera client", zap.Error(err))
	}

	apiKeyStore := limiter.NewAPIKeyStore(viper.Get("apiKeys"))
	tieredLimiter := limiter.NewTieredLimiter(viper.GetStringMap("limiter"), viper.GetInt("hedera.hbarBudget"))

	mClient := hedera.NewMirrorClient(viper.GetString("mirrorNode.baseUrl"), viper.GetInt("mirrorNode.timeoutSeconds"), log)

	enforceAPIKey := viper.GetBool("features.enforceApiKey")

	router := transport.SetupRouter(hClient, mClient, log, apiKeyStore, tieredLimiter, enforceAPIKey)
	port := viper.GetString("server.port")
	log.Info("Starting Hederium server", zap.String("port", port))
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to run server", zap.Error(err))
	}
}
