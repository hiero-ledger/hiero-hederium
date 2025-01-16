package transport

import (
	"net/http"

	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"github.com/gin-gonic/gin"
)

func AuthAndRateLimitMiddleware(apiKeyStore *limiter.APIKeyStore, tieredLimiter *limiter.TieredLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}

		tier, exists := apiKeyStore.GetTierForKey(apiKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}

		if !tieredLimiter.CheckLimits(apiKey, tier) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Set("apiKey", apiKey)
		c.Set("tier", tier)

		c.Next()
	}
}
