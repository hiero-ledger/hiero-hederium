# Hedera Hardhat Example Project (TypeScript)

This project demonstrates how to use Hardhat with the Hedera JSON RPC Relay using TypeScript. It includes examples of basic operations like deploying contracts, making contract calls, transferring tokens, and estimating gas.

## Project Structure

```
.
├── contracts/
│   └── SimpleStorage.sol      # Example smart contract
├── test/
│   └── SimpleStorage.test.ts  # Test cases for various operations
├── scripts/
│   └── deploy.ts             # Deployment script
├── hardhat.config.ts.example # Example Hardhat configuration
├── .env.example             # Example environment variables
├── tsconfig.json            # TypeScript configuration
└── package.json            # Project dependencies
```

## Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Set up your configuration:
   ```bash
   # Copy the example configuration file
   cp hardhat.config.ts.example hardhat.config.ts
   ```

3. Configure your environment:
   - Edit `hardhat.config.ts` with your:
     - Private keys
     - Network endpoints (if using different from defaults)
     - Other settings as needed

4. Run the test suite:
   ```bash
   npx hardhat test --network testnet
   ```

## Configuration

The project uses a configuration system with support for multiple networks:

- **Local Development**:
  ```typescript
  hedera_local: {
    url: "http://localhost:7546",
    accounts: ["your-private-key"],
    chainId: 298
  }
  ```

## Available Commands

- Compile contracts:
  ```bash
  npx hardhat compile
  ```

- Run tests:
  ```bash
  # Run on testnet
  npx hardhat test --network testnet

  # Run on local node
  npx hardhat test --network hedera_local

  ```

- Deploy contract:
  ```bash
  # Deploy to testnet
  npx hardhat run scripts/deploy.ts --network testnet

  # Deploy to local node
  npx hardhat run scripts/deploy.ts --network hedera_local

  # Deploy to mainnet
  npx hardhat run scripts/deploy.ts --network mainnet
  ```

- Format code:
  ```bash
  npm run format
  ```

## Test Cases

The test suite (`test/SimpleStorage.test.ts`) demonstrates:
- Native token transfers
- Contract deployment
- Contract function calls
- Event emission and verification
- Gas estimation
- Transaction receipt retrieval
- Account operations

## Type Safety

The project uses TypeScript for enhanced type safety and developer experience:
- Full type support for contract interactions via TypeChain
- Type-safe event handling and parameter validation
- Autocomplete support for contract methods and properties
- Compile-time checking for contract interactions

## Security Notes

1. Never commit your `hardhat.config.ts` with real private keys to source control
2. The example private key in configuration files is for demonstration only
3. Always use separate private keys for different networks
4. Review and audit contracts before deploying to mainnet

## Troubleshooting

1. If you get TypeScript errors about missing types:
   ```bash
   npx hardhat compile  # This will generate contract types
   ```

2. If you get network connection errors:
   - Ensure your JSON-RPC endpoint is accessible
   - Verify your private key has sufficient funds
   - Check the network's chain ID matches the configuration

3. Common issues:
   - Make sure you've copied and configured `hardhat.config.ts`
   - Ensure you're specifying the correct network when running commands
   - Check that your private key has the correct format (0x prefix if needed)
