package startup

import (
	"fmt"
	"strings"

	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/config"
)

const logo = `
██╗  ██╗███████╗██████╗ ███████╗██████╗ ██╗██╗   ██╗███╗   ███╗
██║  ██║██╔════╝██╔══██╗██╔════╝██╔══██╗██║██║   ██║████╗ ████║
███████║█████╗  ██║  ██║█████╗  ██████╔╝██║██║   ██║██╔████╔██║
██╔══██║██╔══╝  ██║  ██║██╔══╝  ██╔══██╗██║██║   ██║██║╚██╔╝██║
██║  ██║███████╗██████╔╝███████╗██║  ██║██║╚██████╔╝██║ ╚═╝ ██║
╚═╝  ╚═╝╚══════╝╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝
`

// LogStartup prints the banner and the settings the server starts with. It
// prints no secrets: the operator key is left out and API keys are masked.
func LogStartup(cfg *config.Config) {
	// Print ASCII logo
	fmt.Print(logo)

	// Print application version
	fmt.Printf("Starting Hederium v%s\n\n", cfg.Application.Version)

	// Print network configuration
	fmt.Println("Network Configuration:")
	fmt.Printf("  Network: %s\n", cfg.Hedera.Network)
	fmt.Printf("  Chain ID: %s\n\n", cfg.Hedera.ChainID)

	// Print server configuration
	fmt.Println("Server Configuration:")
	fmt.Printf("  Port: %d\n", cfg.Server.Port)
	fmt.Printf("  Enforce API Key: %v\n", cfg.Features.EnforceAPIKey)
	fmt.Printf("  Enable Batch Requests: %v\n\n", cfg.Features.EnableBatchRequests)

	// Print cache configuration
	fmt.Println("Cache Configuration:")
	fmt.Printf("  Default Expiration: %s\n", cfg.Cache.DefaultExpiration)
	fmt.Printf("  Cleanup Interval: %s\n\n", cfg.Cache.CleanupInterval)

	// Print mirror node configuration
	fmt.Println("Mirror Node Configuration:")
	fmt.Printf("  Base URL: %s\n", cfg.MirrorNode.BaseURL)
	fmt.Printf("  Timeout: %s\n\n", cfg.MirrorNode.Timeout)

	// Print rate limiting configuration
	fmt.Println("Rate Limiting Configuration:")
	fmt.Printf("  HBAR Budget: %d\n\n", cfg.Hedera.HbarBudget)

	// Print API keys if present
	if cfg.APIKeys != nil {
		fmt.Println("API Keys Configuration:")
		apiKeysSlice, ok := cfg.APIKeys.([]interface{})
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
