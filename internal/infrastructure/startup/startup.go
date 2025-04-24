package startup

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const logo = `
██╗  ██╗███████╗██████╗ ███████╗██████╗ ██╗██╗   ██╗███╗   ███╗
██║  ██║██╔════╝██╔══██╗██╔════╝██╔══██╗██║██║   ██║████╗ ████║
███████║█████╗  ██║  ██║█████╗  ██████╔╝██║██║   ██║██╔████╔██║
██╔══██║██╔══╝  ██║  ██║██╔══╝  ██╔══██╗██║██║   ██║██║╚██╔╝██║
██║  ██║███████╗██████╔╝███████╗██║  ██║██║╚██████╔╝██║ ╚═╝ ██║
╚═╝  ╚═╝╚══════╝╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝
`

func LogStartup() {
	// Print ASCII logo
	fmt.Println(logo)

	// Print application version
	version := viper.GetString("application.version")
	fmt.Printf("Starting Hederium v%s\n\n", version)

	// Print network configuration
	network := viper.GetString("hedera.network")
	chainId := viper.GetString("hedera.chainId")
	fmt.Println("Network Configuration:")
	fmt.Printf("  Network: %s\n", network)
	fmt.Printf("  Chain ID: %s\n\n", chainId)

	// Print server configuration
	port := viper.GetString("server.port")
	enforceAPIKey := viper.GetBool("features.enforceApiKey")
	enableBatchRequests := viper.GetBool("features.enableBatchRequests")
	fmt.Println("Server Configuration:")
	fmt.Printf("  Port: %s\n", port)
	fmt.Printf("  Enforce API Key: %v\n", enforceAPIKey)
	fmt.Printf("  Enable Batch Requests: %v\n\n", enableBatchRequests)

	// Print cache configuration
	cacheExpiration := viper.GetDuration("cache.defaultExpiration")
	cacheCleanup := viper.GetDuration("cache.cleanupInterval")
	fmt.Println("Cache Configuration:")
	fmt.Printf("  Default Expiration: %s\n", cacheExpiration)
	fmt.Printf("  Cleanup Interval: %s\n\n", cacheCleanup)

	// Print mirror node configuration
	mirrorNodeURL := viper.GetString("mirrorNode.baseUrl")
	mirrorNodeTimeout := viper.GetInt("mirrorNode.timeoutSeconds")
	fmt.Println("Mirror Node Configuration:")
	fmt.Printf("  Base URL: %s\n", mirrorNodeURL)
	fmt.Printf("  Timeout: %d seconds\n\n", mirrorNodeTimeout)

	// Print rate limiting configuration
	hbarBudget := viper.GetInt("hedera.hbarBudget")
	fmt.Println("Rate Limiting Configuration:")
	fmt.Printf("  HBAR Budget: %d\n\n", hbarBudget)

	// Print API keys if present
	apiKeys := viper.Get("apiKeys")
	if apiKeys != nil {
		fmt.Println("API Keys Configuration:")
		apiKeysSlice, ok := apiKeys.([]interface{})
		if ok {
			for i, key := range apiKeysSlice {
				if keyStr, ok := key.(string); ok {
					fmt.Printf("  Key %d: %s\n", i+1, strings.Repeat("*", len(keyStr)))
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("Starting server...")
}
