# Testing Guide

## Quick Start

```bash
# Run all tests
./run_tests.sh all

# Run quick tests (unit + edge cases)
./run_tests.sh quick

# Run with coverage
./run_tests.sh coverage
```

## Test Script Usage

The `run_tests.sh` script provides comprehensive test execution with multiple modes:

### Test Modes

| Mode | Description | Command |
|------|-------------|---------|
| **all** | Complete test suite (coverage + race + benchmarks) | `./run_tests.sh all` |
| **quick** | Fast tests (unit + edge cases) | `./run_tests.sh quick` |
| **unit** | Unit tests only | `./run_tests.sh unit` |
| **integration** | Integration tests (requires DB/Redis) | `./run_tests.sh integration` |
| **service** | Service layer tests | `./run_tests.sh service` |
| **handler** | HTTP handler tests | `./run_tests.sh handler` |
| **middleware** | Middleware tests | `./run_tests.sh middleware` |
| **edge** | Edge case tests | `./run_tests.sh edge` |
| **load** | Load/performance tests | `./run_tests.sh load` |
| **race** | Race condition detection | `./run_tests.sh race` |
| **bench** | Benchmarks | `./run_tests.sh bench` |
| **coverage** | Tests with coverage report | `./run_tests.sh coverage` |

### Features

✅ **Colored Output** - Easy-to-read test results  
✅ **Coverage Reports** - HTML and terminal coverage  
✅ **Race Detection** - Concurrent safety checks  
✅ **Benchmarks** - Performance measurements  
✅ **Test Statistics** - Detailed pass/fail counts  
✅ **Prerequisite Checks** - Validates environment  

### Coverage Reports

After running tests with coverage, view the HTML report:

```bash
# Generate coverage
./run_tests.sh coverage

# Open in browser
open coverage/coverage.html  # macOS
xdg-open coverage/coverage.html  # Linux
```

## Test Structure

```
tests/
├── unit_test.go           # Core business logic tests
├── integration_test.go    # Database/Redis integration
├── user_service_test.go   # User service tests
├── order_service_test.go  # Order service tests
├── user_test.go           # User handler tests
├── order_test.go          # Order handler tests
├── middleware_test.go     # Middleware tests
├── edge_case_test.go      # Edge cases & validation
├── load_test.go           # Performance tests
└── benchmark_test.go      # Benchmarks
```

## Manual Testing

```bash
# Run specific test file
go test -v ./tests/unit_test.go

# Run with race detector
go test -race ./tests/...

# Run benchmarks
go test -bench=. ./tests/benchmark_test.go

# Run with coverage
go test -cover ./tests/...
```

## CI/CD Integration

```yaml
# Example GitHub Actions
- name: Run Tests
  run: ./run_tests.sh all
```

## Test Coverage Goals

- **Target**: 85%+ coverage
- **Current**: Check with `./run_tests.sh coverage`
- **Critical Paths**: 100% coverage required
