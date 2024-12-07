# K6 Load Testing for Hederium

This directory contains k6 scripts and related configuration for load testing the Hederium JSON-RPC endpoints.

## Overview

- **Scripts Directory (`./scripts/`):** Contains individual test scripts for various JSON-RPC methods (e.g. `eth_blockNumber_test.js`).
- **Common Utilities (`./common.js`):** Provides shared configuration and helper functions for all test scripts.
- **Scenarios Directory (`./scenarios/`):** (Optional) Contains advanced scenario configurations for complex load patterns.
- **Environment Variables:** Used to control endpoint URLs, API keys, and load parameters without changing scripts.

## Prerequisites

- [k6](https://k6.io/) installed on your machine.
- The Hederium service running locally or accessible at the configured URL.

## Configuration

You can control the behavior of the tests using environment variables:

- **ENDPOINT_URL:** The URL of your JSON-RPC endpoint. Default is `http://localhost:7546`.
- **API_KEY:** The API key to send in the `X-API-KEY` header, if needed. Leave empty if not required.
- **VUS:** Number of Virtual Users (VUs) to simulate. Default is `1`.
- **DURATION:** Test duration (e.g. `10s`, `30s`, `1m`). Default is `10s`.

Example:

```bash
ENDPOINT_URL=http://localhost:7546 \
API_KEY=FREE-USER-API-KEY-123 \
VUS=10 \
DURATION=30s \
k6 run ./scripts/eth_blockNumber_test.js
```
