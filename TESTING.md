# Testing Guide

This document describes the testing infrastructure for webfetch-clean.

## Test Suite Overview

The project includes comprehensive test coverage across all core components:

### Test Files

1. **fetcher_test.go** - HTTP client tests
   - Valid URL fetching
   - Error handling (404, 500, timeouts, empty content)
   - HTTP headers verification
   - Network error handling

2. **cleaner_test.go** - HTML cleaning tests
   - Script, style, nav removal
   - Ad detection and removal
   - Clutter removal (sidebar, footer, popups, modals, cookies)
   - Iframe removal
   - `preserveMainOnly` option
   - `removeImages` option
   - Attribute stripping
   - Semantic element preservation

3. **converter_test.go** - Format conversion tests
   - HTML to Markdown conversion
   - Heading, link, list, code conversion
   - Image syntax conversion
   - HTML format passthrough
   - Invalid format handling

4. **integration_test.go** - End-to-end tests
   - Full pipeline (Fetch → Clean → Convert)
   - Complex page cleaning
   - Option combinations
   - Real-world scenarios

## Running Tests

### Run all tests
```bash
go test -v ./...
```

### Run tests with coverage
```bash
go test -v -coverprofile="coverage.out" -covermode=atomic .
go tool cover -html="coverage.out"  # View HTML report
go tool cover -func="coverage.out"  # View function coverage
```

### Run specific test
```bash
go test -v -run TestCleanHTML_RemoveAds
```

### Run tests with race detection
```bash
go test -v -race .
```

## Coverage Report

Current coverage: **40.4%**

Coverage by component:
- **fetcher.go**: 88.0% - Core HTTP fetching logic
- **cleaner.go**: 89.8% - HTML cleaning pipeline
- **converter.go**: 85.7-100% - Format conversion
- **main.go**: 0.0% - CLI/MCP entry points (not covered)

The main.go file has 0% coverage as it contains CLI and MCP server entry points that require integration testing. Core business logic has excellent test coverage.

## GitHub Actions CI/CD

The project includes automated testing via GitHub Actions:

### Workflow: `.github/workflows/test.yml`

**Jobs:**

1. **test** - Run tests on multiple platforms
   - Platforms: Ubuntu, macOS
   - Go versions: 1.21, 1.22, 1.23
   - Runs tests with race detection
   - Generates coverage reports
   - Uploads coverage to Codecov

2. **lint** - Code quality checks
   - Runs golangci-lint
   - Configuration: `.golangci.yml`

3. **build** - Build verification
   - Builds binary on multiple platforms
   - Tests binary execution
   - Uploads build artifacts

### Triggers

- Push to main/develop branches
- Pull requests to main/develop branches
- Manual workflow dispatch

### Configuration

Linting configuration is in `.golangci.yml` with enabled linters:
- errcheck, gosimple, govet, ineffassign, staticcheck, unused
- gofmt, goimports, misspell, revive

## Known Limitations

### Ad Detection

The ad detection is intentionally aggressive. It matches patterns like:
- `advertisement`
- `ad-`, `-ad-`, `-ad`
- `_ad_`, `_ad`, `ad_`
- ` ad ` (space-surrounded)

This means legitimate content with "ad" in class names may be removed:
- `thread-card` matches `-ad-` pattern
- `reader-mode` matches `-ad-` pattern

This is acceptable because:
1. Better to be overly aggressive in removing ads
2. Most legitimate content doesn't have "ad" surrounded by dashes
3. Main semantic content is usually not affected

## Test Maintenance

### Adding New Tests

1. Create test function with descriptive name: `TestComponentName_Behavior`
2. Use table-driven tests for similar scenarios
3. Include edge cases and error conditions
4. Document complex test scenarios with comments

### Testing Best Practices

1. **Arrange-Act-Assert**: Structure tests clearly
2. **Descriptive names**: Test names should describe what they test
3. **Error messages**: Include helpful error messages in assertions
4. **Independence**: Tests should not depend on each other
5. **Fast tests**: Keep tests fast by using mocks/fakes where appropriate

## Debugging Tests

### View test output
```bash
go test -v .
```

### Debug specific test
```bash
go test -v -run TestName 2>&1 | less
```

### Add debug output
```go
t.Logf("Debug: variable = %v", variable)
```

## Future Improvements

Potential testing enhancements:

1. **MCP Protocol Tests**: Add tests for MCP JSON-RPC handlers
2. **Benchmark Tests**: Add performance benchmarks for cleaning pipeline
3. **Fuzzing**: Add fuzz tests for HTML parsing edge cases
4. **E2E Tests**: Add tests that fetch real URLs (optional, slow)
5. **Coverage Goal**: Aim for >80% coverage by testing CLI/MCP handlers

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Coverage Documentation](https://go.dev/blog/coverage)
