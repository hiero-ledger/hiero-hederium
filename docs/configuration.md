# Configuration

This document describes all available configuration options for the Hederium service. The configuration can be provided through a YAML file located at `configs/config.yaml`.

## Overview

The configuration is structured into several main sections:
- Environment
- Application
- Server
- Hedera
- Mirror Node
- Rate Limiter
- Logging
- API Keys
- Features
- Cache

## Configuration Options

| Section | Option | Type | Default Value | Description |
|---------|--------|------|---------------|-------------|
| `environment` | - | string | `"development"` | Application environment setting |
| **Application** |
| `application.version` | - | string | `"0.1.0"` | Version of the application |
| **Server** |
| `server.port` | - | integer | `7546` | HTTP server port |
| **Hedera** |
| `hedera.network` | - | string | `"testnet"` | Hedera network to connect to |
| `hedera.operatorId` | - | string | `"0.0.1466"` | Hedera operator account ID |
| `hedera.operatorKey` | - | string | - | Hedera operator private key |
| `hedera.chainId` | - | string | `"0x128"` | Chain ID in hexadecimal format |
| `hedera.hbarBudget` | - | integer | `1000` | HBAR budget limit |
| **Mirror Node** |
| `mirrorNode.baseUrl` | - | string | `"https://testnet.mirrornode.hedera.com"` | Base URL for the Hedera Mirror Node |
| `mirrorNode.timeoutSeconds` | - | integer | `10` | Timeout for mirror node requests |
| **Rate Limiter** |
| `limiter.free.requestsPerMinute` | - | integer | `100` | Request limit per minute for free tier |
| `limiter.free.hbarLimit` | - | integer | `10` | HBAR limit for free tier |
| `limiter.premium.requestsPerMinute` | - | integer | `1000` | Request limit per minute for premium tier |
| `limiter.premium.hbarLimit` | - | integer | `10000` | HBAR limit for premium tier |
| **Logging** |
| `logging.level` | - | string | `"debug"` | Log level (debug, info, warn, error) |
| `logging.DisableCaller` | - | boolean | `true` | Disable caller information in logs |
| **API Keys** |
| `apiKeys` | - | array | - | List of API keys and their tiers |
| **Features** |
| `features.enforceApiKey` | - | boolean | `false` | Enable/disable API key enforcement |
| **Cache** |
| `cache.defaultExpiration` | - | duration | `"1h"` | Default cache entry expiration time |
| `cache.cleanupInterval` | - | duration | `"30m"` | Cache cleanup interval |

## Example Configuration

```yaml
environment: "development"

application:
  version: "0.1.0"

server:
  port: 7546

hedera:
  network: "testnet"
  operatorId: "0.0.1466"
  operatorKey: "your-operator-key"
  chainId: "0x128"
  hbarBudget: 1000

mirrorNode:
  baseUrl: "https://testnet.mirrornode.hedera.com"
  timeoutSeconds: 10

limiter:
  free:
    requestsPerMinute: 100
    hbarLimit: 10
  premium:
    requestsPerMinute: 1000
    hbarLimit: 10000

logging:
  level: "debug"
  DisableCaller: true

apiKeys:
  - key: "FREE-USER-API-KEY-123"
    tier: "free"
  - key: "PREMIUM-USER-API-KEY-456"
    tier: "premium"

features:
  enforceApiKey: false

cache:
  defaultExpiration: "1h"
  cleanupInterval: "30m"
```

## Notes

- The `hedera.operatorKey` should be kept secure and not shared publicly
- Duration values (like cache settings) support Go duration format (e.g., "1h", "30m", "24h")
- Log levels supported: "debug", "info", "warn", "error"
- API keys should be properly secured and not committed to version control
- Chain ID must be provided in hexadecimal format
