# Hederium

<p align="center">
  <img src="assets/logo.png" alt="Hederium Logo" width="100%"/>
</p>

[![CI - Unit Tests](https://github.com/LimeChain/Hederium/actions/workflows/unit-test.yml/badge.svg)](https://github.com/LimeChain/Hederium/actions/workflows/unit-test.yml)
[![CI - Dev Tool Tests](https://github.com/LimeChain/Hederium/actions/workflows/dev-tool-test.yml/badge.svg?branch=main)](https://github.com/LimeChain/Hederium/actions/workflows/dev-tool-test.yml)
[![License](https://img.shields.io/badge/license-apache2-blue.svg)](LICENSE)
[![Made With](https://img.shields.io/badge/made_with-go-blue)](https://golang.org/)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/hiero-ledger/hiero-hederium/badge)](https://scorecard.dev/viewer/?uri=github.com/hiero-ledger/hiero-hederium)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/10697/badge)](https://bestpractices.coreinfrastructure.org/projects/10697)

Hederium is designed to implement the Hashio protocol in Go, offering high performance, reliability, and scalability. This implementation follows Go best practices and provides a comprehensive suite of features for building distributed applications.

## Prerequisites

- Go 1.20 or higher
- Docker (for containerized deployment)
- Git

## Documentation

- [Configuration](docs/configuration.md)
- [RPC API](docs/rpc-api.md)
- [Benchmark](docs/benchmark.md)

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/LimeChain/Hederium.git
   cd Hederium
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Building and Running

### Local Development

```bash
go run ./cmd/server/main.go
```

### Using Docker

```bash
docker build -t hederium .
docker run hederium
```

## Testing

Run the unit tests:

```bash
go test ./... -v
```

## Project Structure

- `/cmd` - Main applications
- `/internal` - Private application and library code
- `/configs` - Configuration files
- `/test` - Additional external test applications and test data
- `/tools` - Example tools for testing and development
- `/docs` - Design and user documents

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please read our [contributing guide](https://github.com/hiero-ledger/.github/blob/main/CONTRIBUTING.md) to see how you can get involved.

## Code of Conduct

Hiero uses the Linux Foundation Decentralised Trust [Code of Conduct](https://www.lfdecentralizedtrust.org/code-of-conduct).

## License

This project is licensed under the terms of the LICENSE file included in the repository.

## Contact

Project Link: [https://github.com/LimeChain/Hederium](https://github.com/LimeChain/Hederium).
Join our [community discussions](https://discord.lfdecentralizedtrust.org/) on discord.
Attend our [community calls](https://zoom-lfx.platform.linuxfoundation.org/meetings/hiero?view=week). 
