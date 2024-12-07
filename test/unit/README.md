# Unit Tests for Hederium

This directory contains the unit tests for the Hederium application. Unit tests focus on individual components (functions, methods, small modules) in isolation, using mocks or stubs to replace external dependencies.

## Directory Structure

```
test/unit/
├── mocks/          # Generated mock files using gomock
├── service/        # Unit tests for internal/service package
├── infrastructure/ # Unit tests for internal/infrastructure package
└── ...            # More subdirectories mirroring internal/ structure
```

The `unit` directory mirrors the `internal` directory structure of the application, making it easy to locate corresponding test files for each source file.

## Adding New Unit Tests

1. **Identify Where the Test Belongs:**
   Place the test file in a directory structure that matches the package under test. For example:

   - If you’re testing `internal/service/eth_service.go`, create or use `test/unit/service/eth_service_test.go`.
   - If the directory doesn’t exist, create it.

2. **Name the Test File Appropriately:**
   Test files follow the `*_test.go` naming pattern. For example:  
   `eth_service_test.go`

3. **Use the `_test` Suffix for Package Names (Optional):**
   To avoid import cycles and clarify the test context, you can name the package `<packagename>_test`. For example:
   ```go
   package service_test
   ```
4. **Write Your Tests: Use the Go testing package and a testing framework like testify for assertions:**

   ```go
   import (
   "testing"
   "github.com/stretchr/testify/assert"
   )

   func TestExample(t \*testing.T) {
   result := SomeFunctionUnderTest()
       assert.Equal(t, expected, result)
   }
   ```

## Generating and Regenerating Mocks

If your tests depend on interfaces and mocks:

When to Generate Mocks:
If you introduce a new interface or change an existing one (add methods, remove methods), you must regenerate the mocks to keep them in sync.

How to Generate Mocks: Use mockgen against the interface in your application code:

```bash
mockgen -source=internal/infrastructure/hedera/mirror_client.go \
 -destination=test/unit/mocks/mock_mirror_client.go \
 -package=mocks \
 -mock_names=MirrorNodeClient=MockMirrorClient
```

Adjust paths and package names as needed.

After Changing Interfaces: If the interface changes, rerun the mockgen command. If you encounter compilation errors in tests related to mock methods, it likely means the interface and mock are out of sync.

## Running the Tests

To run all unit tests:

```bash
go test ./test/unit/... -v
```

Options:

-v for verbose output.
-count=1 to avoid test caching:

```bash
go test ./test/unit/... -v -count=1
```

To run a specific package or test file:

```bash
go test ./test/unit/service -v
go test ./test/unit/service/eth_service_test.go -v
```

## Test Coverage

To check test coverage:

```bash
go test ./test/unit/... -cover
```

To generate an HTML coverage report:

```bash
go test ./test/unit/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Open the generated report in your browser to see which lines of code are covered by tests.

## Best Practices

- Keep Tests Independent:
  Each test should run independently of others. Mocks and other dependencies should be set up and torn down within each test.

- Focus on One Responsibility per Test:
  Write tests that focus on a single aspect of the function or method under test. This makes failures easier to diagnose.

- Update Tests as Code Changes: Whenever you refactor or add new logic to your code, update or add tests to maintain confidence and prevent regressions.

## Troubleshooting

- If tests fail due to nil pointers, ensure all required dependencies (like logger or mClient in EthService) are properly initialized in the test.
- If you see command not found: mockgen, ensure your PATH includes $(go env GOPATH)/bin or run go install again.
- If your mocks are out of date, regenerate them with mockgen.
