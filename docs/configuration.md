# Configuration

Hederium reads its settings once, at start-up, from environment variables and an optional YAML file. If a setting is missing or has a bad value, the server does not start and prints every problem it found.

## Where a setting comes from

Each setting is looked up in four places. The first one that has a value wins:

1. **The relay's environment variable.** Hederium accepts the names the [Hedera JSON-RPC relay](https://github.com/hiero-ledger/hiero-json-rpc-relay) uses, such as `MIRROR_NODE_URL`, so a relay deployment can reuse its environment.
2. **Hederium's own environment variable.** This is the YAML key in upper case with `_` in place of `.`, such as `MIRRORNODE_BASEURL`. The older dotted form, `MIRRORNODE.BASEURL`, still works.
3. **The file `configs/config.yaml`**, relative to the working directory. The file is optional.
4. **The default.**

An environment variable that is set but empty counts as not set.

## Settings

| Relay variable | Hederium variable | YAML key | Default | Description |
| --- | --- | --- | --- | --- |
| `HEDERA_NETWORK` | `HEDERA_NETWORK` | `hedera.network` | none, mandatory | Network to connect to: `mainnet`, `testnet`, `previewnet` or `local` |
| `CHAIN_ID` | `HEDERA_CHAINID` | `hedera.chainId` | none, mandatory | Chain ID, in decimal (`296`) or hex (`0x128`) |
| `MIRROR_NODE_URL` | `MIRRORNODE_BASEURL` | `mirrorNode.baseUrl` | none, mandatory | Base URL of the Mirror Node REST API (`http` or `https`) |
| `OPERATOR_ID_MAIN` | `HEDERA_OPERATORID` | `hedera.operatorId` | none, mandatory | Operator account ID |
| `OPERATOR_KEY_MAIN` | `HEDERA_OPERATORKEY` | `hedera.operatorKey` | none, mandatory | Operator private key. Never printed or logged |
| `MIRROR_NODE_TIMEOUT` | `MIRRORNODE_TIMEOUTSECONDS` | `mirrorNode.timeoutSeconds` | 10 seconds | Timeout of a Mirror Node request, above zero. See [Units](#units) |
| `SERVER_PORT` | `SERVER_PORT` | `server.port` | `7546` | HTTP port, 1 to 65535 |
| `LOG_LEVEL` | `LOGGING_LEVEL` | `logging.level` | `trace` | `trace`, `debug`, `info`, `warn`, `error`, `fatal` or `silent`. See [Log levels](#log-levels) |
| `BATCH_REQUESTS_ENABLED` | `FEATURES_ENABLEBATCHREQUESTS` | `features.enableBatchRequests` | `true` | Accept JSON-RPC batch requests |
| `CACHE_TTL` | `CACHE_DEFAULTEXPIRATION` | `cache.defaultExpiration` | 1 hour | Default lifetime of a cache entry. Zero means entries never expire. See [Units](#units) |
| | `CACHE_CLEANUPINTERVAL` | `cache.cleanupInterval` | `30m` | How often expired cache entries are removed. Zero turns the clean-up off |
| | `FEATURES_ENFORCEAPIKEY` | `features.enforceApiKey` | `false` | Require an `X-API-KEY` header on every request |
| | `APPLICATION_VERSION` | `application.version` | `0.1.0` | Version reported by `web3_clientVersion` |
| | `HEDERA_HBARBUDGET` | `hedera.hbarBudget` | `0` | Operator HBAR budget. Not enforced yet |

Booleans are `true` or `false`, in any letter case. Any other value is refused.

The API-key settings, `apiKeys` and `limiter`, are lists and maps, so they can only be set in the YAML file:

| YAML key | Description |
| --- | --- |
| `apiKeys` | List of `key` and `tier` pairs |
| `limiter.<tier>.requestsPerMinute` | Requests per minute for one API key of that tier |
| `limiter.<tier>.hbarLimit` | HBAR limit for one API key of that tier |

### Units

A duration's unit depends on the name it is set through:

| Setting | Relay variable | Hederium variable and YAML key |
| --- | --- | --- |
| Mirror Node timeout | `MIRROR_NODE_TIMEOUT`: whole milliseconds, `10000` | `MIRRORNODE_TIMEOUTSECONDS`, `mirrorNode.timeoutSeconds`: whole seconds, `10` |
| Cache lifetime | `CACHE_TTL`: whole milliseconds, `3600000` | `CACHE_DEFAULTEXPIRATION`, `cache.defaultExpiration`: Go duration, `1h` |
| Cache clean-up | | `CACHE_CLEANUPINTERVAL`, `cache.cleanupInterval`: Go duration, `30m` |

A Go duration is a number with a unit, such as `90s`, `30m` or `1h`.

### Log levels

Hederium logs with zap, which has no `trace` or `silent` level. `trace` runs as `debug`, and `silent` runs as `fatal`, the quietest level zap has.

## Differences from the relay

- Only the relay variables in the table above are read. Other relay variables are ignored.
- The relay reads a boolean other than `true` as false, and a `CHAIN_ID` that is not a number as `0xNaN`. Hederium refuses both at start-up.
- Hederium has no read-only mode, so the operator ID and key are always mandatory.
- Hederium does not read a `.env` file. Pass the variables through the environment, for example with `docker run --env-file`.

## Examples

Environment only, with relay names:

```bash
export HEDERA_NETWORK=testnet
export CHAIN_ID=0x128
export MIRROR_NODE_URL=https://testnet.mirrornode.hedera.com
export OPERATOR_ID_MAIN=0.0.1234
export OPERATOR_KEY_MAIN=<operator private key>
./hederium
```

The same settings in `configs/config.yaml`:

```yaml
application:
  version: "0.1.0"

server:
  port: 7546

hedera:
  network: "testnet"
  operatorId: "0.0.1234"
  operatorKey: "<operator private key>"
  chainId: "0x128"

mirrorNode:
  baseUrl: "https://testnet.mirrornode.hedera.com"
  timeoutSeconds: 10

logging:
  level: "info"

features:
  enforceApiKey: false
  enableBatchRequests: true

cache:
  defaultExpiration: "1h"
  cleanupInterval: "30m"
```

## Notes

- Keep the operator key and API keys out of version control. Prefer the environment, or a secret store that sets it, over the YAML file.
- Error messages name the setting that is wrong and never repeat its value.
