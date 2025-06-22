# Test Directory Structure

This directory contains organized test files and utilities for the KoeMoji-Go project.

## Directory Structure

```
test/
├── README.md                    # This file
├── coverage.out                 # Coverage data files
├── coverage.html               # Coverage reports  
├── integration_test.go         # Integration tests
├── koemoji-go-debug           # Debug binaries
├── koemoji-go-test            # Test binaries
├── shared/                      # Shared test utilities
│   ├── common.go               # Common test functions and config
│   ├── helpers.go              # Additional test helpers
│   ├── performance.go          # Performance testing utilities
│   └── testutil.go             # Test utilities
├── testdata/                   # Test data files
│   └── test_recording_1.wav    # Sample audio files
├── manual-test-commands.md     # Manual testing commands
└── recorder_manual.go          # Manual recorder testing

internal/
├── config/
│   ├── config_test.go          # Main config tests
│   ├── configure_test.go       # Configuration UI tests
│   └── testdata/
│       └── helpers.go          # Config-specific test helpers
├── whisper/
│   ├── whisper_test.go         # Main whisper tests
│   └── testdata/
│       ├── mocks_test.go       # Whisper mocks
│       └── test_helpers.go     # Whisper-specific helpers
├── llm/
│   ├── llm_test.go             # Main LLM tests
│   └── testdata/
│       ├── http_mocks.go       # HTTP client mocks
│       └── test_helpers.go     # LLM-specific helpers
├── logger/
│   └── logger_test.go          # Logger tests
├── processor/
│   └── processor_test.go       # Processor tests
└── recorder/
    └── recorder_test.go        # Recorder tests
```

## Test Organization Principles

### 1. Package-specific testdata directories
- Each package has its own `testdata/` directory for package-specific test utilities
- Mocks and helpers are organized by functionality

### 2. Shared utilities
- Common test functions are in `test/shared/`
- Avoid duplication across packages

### 3. Naming conventions
- `*_test.go` - Main test files
- `testdata/*.go` - Test helpers and mocks
- `mock*.go` - Mock implementations
- `*helpers.go` - Test helper functions

## Test Coverage Goals

| Package | Target Coverage | Current Status |
|---------|----------------|----------------|
| config  | 80%           | ✅ 73.8%      |
| whisper | 70%           | ✅ 71.8%      |
| llm     | 75%           | ✅ 69.4%      |
| logger  | 95%           | ✅ 100%       |
| processor | 70%         | 🔄 21.6%      |
| recorder | 70%          | 🔄 41.6%      |
| gui     | 40%           | ❌ 0%         |
| cmd     | 60%           | ❌ 0%         |

## Running Tests

### All tests
```bash
go test ./... -v -cover
```

### Specific package
```bash
go test ./internal/config -v -cover
go test ./internal/whisper -v -cover
go test ./internal/llm -v -cover
```

### With timeout (for tests that might hang)
```bash
go test ./... -v -cover -timeout 30s
```

### Generate coverage report
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Test Types

### 1. Unit Tests
- Test individual functions in isolation
- Use mocks for external dependencies
- Fast execution

### 2. Integration Tests
- Test component interactions
- May use real file system or network (with timeouts)
- Focus on end-to-end workflows

### 3. Benchmark Tests
- Performance testing for critical paths
- Located alongside unit tests with `Benchmark*` prefix

## Mock Strategy

### HTTP Mocks (LLM package)
- `MockHTTPClient` for API testing
- Predefined responses for common scenarios

### Command Mocks (Whisper package)
- `MockCommandExecutor` for external command testing
- File system mocks for path operations

### Reader Mocks (Config package)
- `MockReader` for interactive input testing
- Simulates user input for configuration UI

## Best Practices

1. **Use t.Helper()** in test helper functions
2. **Create isolated test environments** with t.TempDir()
3. **Mock external dependencies** (APIs, commands, file system)
4. **Test error conditions** as well as success paths
5. **Use table-driven tests** for multiple test cases
6. **Add benchmarks** for performance-critical code
7. **Keep tests deterministic** and independent
8. **Use descriptive test names** that explain the scenario