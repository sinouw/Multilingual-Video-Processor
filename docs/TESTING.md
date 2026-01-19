# Testing Guide

This document provides guidelines for writing and running tests for the Multilingual Video Processor.

## Running Tests

### Run All Tests

```bash
go test ./...
```

### Run Tests with Verbose Output

```bash
go test -v ./...
```

### Run Tests with Coverage

```bash
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Or use the Makefile:

```bash
make test-coverage
```

### Run Specific Test Package

```bash
go test ./internal/api
go test ./internal/storage
```

### Run Specific Test Function

```bash
go test -v ./internal/api -run TestHealthHandler
```

## Test Structure

Tests follow Go's standard testing conventions:

- Test files: `*_test.go`
- Test functions: `TestFunctionName(t *testing.T)`
- Test cases: Use table-driven tests where appropriate

### Example Test

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        {
            name:     "invalid input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("FunctionName() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Test Categories

### Unit Tests

Unit tests test individual functions and components in isolation:

- Location: Same package as code being tested
- File naming: `*_test.go`
- Examples: `internal/api/health_test.go`, `internal/validator/request_test.go`

### Integration Tests

Integration tests test component interactions:

- Location: `test/integration/`
- Build tag: `// +build integration`
- Run with: `go test -tags=integration ./test/integration/`
- Note: Set `RUN_INTEGRATION_TESTS=1` environment variable

## Mocking External Services

For testing components that depend on external services:

### Strategy 1: Use Interfaces

Define interfaces for external dependencies:

```go
type SpeechToTextClient interface {
    Recognize(ctx context.Context, req *speechpb.RecognizeRequest) (*speechpb.RecognizeResponse, error)
}
```

### Strategy 2: Mock HTTP Servers

For HTTP-based APIs, use `httptest`:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Mock response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(mockResponse)
}))
defer server.Close()

// Use server.URL as API endpoint
```

### Strategy 3: Test Doubles

Create test doubles for complex dependencies:

```go
type mockStorage struct {
    files map[string][]byte
}

func (m *mockStorage) Download(ctx context.Context, bucket, path string) (string, error) {
    // Mock implementation
}
```

## Context Testing

All functions that accept `context.Context` should be tested with:

- Cancelled contexts
- Timed-out contexts
- Background contexts

Example:

```go
func TestFunction_ContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    err := Function(ctx, "input")
    if err == nil {
        t.Error("expected error for cancelled context")
    }
}
```

## Error Handling Tests

Test error conditions:

- Invalid inputs
- Missing required fields
- Resource not found errors
- Timeout errors
- Cancellation errors

## Test Data

- Use test fixtures in `test/fixtures/` for large files
- Create temporary files in tests, clean up in `defer`
- Use `t.Helper()` in test helper functions

## Coverage Goals

- Minimum coverage: 60%
- Target coverage: 70-80%
- Critical paths: 90%+

## Running Tests in CI

Tests run automatically in CI on:

- Push to `main` or `develop` branches
- Pull requests to `main`
- See `.github/workflows/ci.yml` for configuration

## Environment Variables for Testing

Some tests require environment variables:

```bash
export GCS_BUCKET_OUTPUT=test-bucket
export GOOGLE_TRANSLATE_API_KEY=test-key
export RATE_LIMIT_RPM=100
export JOB_TTL=1h
```

Use `test/helpers.go` for test setup utilities.

## Best Practices

1. **Test Independence**: Each test should be independent and not rely on other tests
2. **Clean State**: Reset state between tests
3. **Fast Tests**: Keep tests fast; use mocks for slow operations
4. **Clear Names**: Use descriptive test names
5. **Error Messages**: Include helpful error messages in assertions
6. **Table-Driven Tests**: Use table-driven tests for multiple test cases
7. **Context Handling**: Always test context cancellation and timeouts
8. **Edge Cases**: Test edge cases (empty strings, nil values, boundary conditions)

## Example Test Files

- `internal/api/health_test.go` - Health endpoint tests
- `internal/api/status_test.go` - Status endpoint tests with mocks
- `internal/validator/request_test.go` - Validation tests
- `internal/utils/uuid_test.go` - Utility function tests

## Continuous Integration

Tests run automatically in GitHub Actions. See `.github/workflows/ci.yml` for:

- Test execution
- Coverage reporting
- Security scanning
- Linting
