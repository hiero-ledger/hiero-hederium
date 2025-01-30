# Ethereum Development Tools Support

The JSON RPC relay serves as an interface to the Hedera network for Ethereum developer tools that utilize the implemented JSON RPC APIs. This bridge enables developers to use familiar Ethereum development tools while working with the Hedera network.

## Overview

Our relay supports various Ethereum development tools, allowing developers to:
- Deploy and interact with smart contracts
- Perform token transfers
- Execute contract calls
- Query blockchain state
- And more...

## Supported Tools

We currently support integration with the following Ethereum development tools. Click on each tool for detailed compatibility information:

- [Hardhat](#hardhat) - Popular development environment for Ethereum
- [ethers.js](#ethersjs) - Coming soon
- [Foundry](#foundry) - Coming soon

## Hardhat

Hardhat is a development environment designed for Ethereum software compilation, deployment, testing, and debugging. Below is our current support status for essential operations:

### Basic Operations
| Operation                               | Status | Notes |
| --------------------------------------- | ------ | ----- |
| HBAR Transfers                          | ❌     | Transfer of native HBAR tokens |
| Contract Deployment                     | ❌     | Smart contract deployment |
| Contract Calls                          | ❌     | Execution of contract functions |
| View/Pure Function Calls                | ❌     | Reading contract state |
| Gas Estimation                          | ❌     | Transaction gas estimation |
| Event Logs                             | ❌     | Contract event monitoring |
| Transaction Receipt Retrieval           | ❌     | Getting transaction details |
| Account Management                      | ❌     | Account balance and nonce operations |

## Important Notes

1. **Rate Limiting**: Development tools often make multiple requests to certain endpoints, especially during contract deployment. Be mindful of rate limiting when deploying multiple large contracts.

2. **Debug Operations**: Debug operations are currently not supported across any tools. This includes:
   - Step-by-step debugging
   - State inspection during execution
   - Call stack examination
   - Variable inspection

3. **Gas Optimization**: While gas estimation is supported, the actual gas usage on Hedera may differ from Ethereum due to network differences.

## Coming Soon

We are actively working on expanding tool support and adding new features:
- Debug operation support
- Additional tool compatibility
- Enhanced testing capabilities
- Improved gas estimation accuracy
- Support verification for all listed tools

For the latest updates and detailed configuration guides, please refer to our documentation.