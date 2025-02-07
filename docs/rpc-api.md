# RPC API Documentation

This document describes the JSON-RPC APIs exposed by the Hederium relay. The relay implements Ethereum JSON-RPC compatible APIs by translating requests between Ethereum and Hedera networks.

## Overview

The relay supports three main categories of APIs:
- `eth_*` - Ethereum-compatible APIs for interacting with the Hedera network
- `net_*` - Network-related APIs
- `web3_*` - Web3-related utilities

## API Methods

Below is a comprehensive list of supported RPC methods and their dependencies on Hedera's Mirror Node and/or Consensus Node services.

| Method | Description | Mirror Node | Consensus Node |
|--------|-------------|-------------|----------------|
| `eth_getBlockByHash` | Gets block information by hash | ✅ | |
| `eth_getBlockByNumber` | Gets block information by number | ✅ | |
| `eth_getBalance` | Gets account balance | ✅ | |
| `eth_getTransactionCount` | Gets account nonce/transaction count | ✅ | |
| `eth_estimateGas` | Estimates gas for transaction | ✅ | |
| `eth_call` | Executes a call without creating a transaction | ✅ | |
| `eth_getTransactionByHash` | Gets transaction details by hash | ✅ | |
| `eth_getTransactionReceipt` | Gets transaction receipt | ✅ | |
| `eth_feeHistory` | Gets historical fee information | ✅ | |
| `eth_getStorageAt` | Gets contract storage at position | ✅ | |
| `eth_getLogs` | Gets event logs matching filter | ✅ | |
| `eth_getBlockTransactionCountByHash` | Gets transaction count in block by hash | ✅ | |
| `eth_getBlockTransactionCountByNumber` | Gets transaction count in block by number | ✅ | |
| `eth_getTransactionByBlockHashAndIndex` | Gets transaction by block hash and index | ✅ | |
| `eth_getTransactionByBlockNumberAndIndex` | Gets transaction by block number and index | ✅ | |
| `eth_sendRawTransaction` | Submits raw transaction | ✅ | ✅ |
| `eth_getCode` | Gets contract code | ✅ | ✅ |
| `eth_blockNumber` | Gets latest block number | ✅ | |
| `eth_gasPrice` | Gets current gas price | ✅ | |
| `eth_chainId` | Gets network chain ID | ✅ | |
| `eth_accounts` | Gets list of accounts (returns empty) | | |
| `eth_syncing` | Gets sync status (always false) | | |
| `eth_mining` | Gets mining status (always false) | | |
| `eth_maxPriorityFeePerGas` | Gets max priority fee (always 0x0) | | |
| `eth_hashrate` | Gets hash rate (always 0x0) | | |
| `eth_getUncleCountByBlockNumber` | Gets uncle count (always 0x0) | | |
| `eth_getUncleByBlockNumberAndIndex` | Gets uncle by block (always null) | | |
| `eth_getUncleCountByBlockHash` | Gets uncle count by hash (always 0x0) | | |
| `eth_getUncleByBlockHashAndIndex` | Gets uncle by hash (always null) | | |
| `net_listening` | Gets network listening status (always false) | | |
| `net_version` | Gets network version | | |
| `web3_clientVersion` | Gets client version | | |

## Notes

1. Most APIs primarily rely on the Mirror Node for data retrieval
2. Only `eth_sendRawTransaction` and `eth_getCode` require both Mirror Node and Consensus Node interaction
3. Some Ethereum APIs are implemented to return constant values for compatibility:
   - `eth_accounts` - Returns empty array
   - `eth_syncing` - Returns false
   - `eth_mining` - Returns false
   - `eth_maxPriorityFeePerGas` - Returns 0x0
   - `eth_hashrate` - Returns 0x0
   - All uncle-related methods return 0x0 or null
4. Network APIs (`net_*`) are minimal implementations:
   - `net_listening` always returns false
   - `net_version` returns the chain ID
5. Web3 API only provides client version information
