# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated testing and quality assurance.

## Workflows

### 1. Simple Tests (`simple-test.yml`)
**Triggers:** Every push and pull request to any branch
**Purpose:** Fast feedback on code changes using root Makefile
**What it does:**
- Uses `make test-all` to run tests for all Go services efficiently
- Uses `make build-all` to build all services
- Runs UI tests and builds the UI
- Single job for faster execution

### 2. Quick Tests (`quick-test.yml`)
**Triggers:** Every push and pull request to any branch
**Purpose:** Fast feedback on code changes with parallel execution
**What it does:**
- Runs unit tests for all Go services using matrix strategy (parallel execution)
- Builds all Go services to ensure compilation
- Runs UI tests and builds the UI
- Uses matrix strategy to run services in parallel

### 3. Full Test Suite (`full-test.yml`)
**Triggers:** Push to main branch and pull requests to main
**Purpose:** Comprehensive testing before merging to main
**What it does:**
- All tests from Quick Tests
- Integration tests with PostgreSQL database
- Security scanning (Go vulnerabilities, npm audit)
- Code quality checks (Go vet, gofmt, ESLint)
- Detailed reporting

### 4. Complete Test Suite (`test.yml`)
**Triggers:** Every push and pull request to any branch
**Purpose:** Most comprehensive testing (legacy workflow)
**What it does:**
- All tests from Full Test Suite
- Additional security and quality checks
- More detailed notifications

## Workflow Features

### Parallel Execution
- Services are tested in parallel using matrix strategy
- UI tests run independently
- Integration tests run after unit tests complete

### Caching
- Go modules are cached to speed up builds
- Node.js dependencies are cached for UI builds

### Database Integration
- PostgreSQL service is provided for integration tests
- Database initialization is automated
- Health checks ensure database is ready before tests

### Security Scanning
- Go vulnerability scanning using `govulncheck`
- npm audit for JavaScript dependencies
- Moderate security level threshold

### Code Quality
- Go vet for static analysis
- gofmt for code formatting consistency
- ESLint for JavaScript code quality

## Local Testing

Before pushing code, you can run tests locally:

```bash
# Test all services
make test-all

# Test specific service
cd data-service && make test

# Test UI
cd ui && npm test

# Build all services
make build-all
```

## Workflow Status

You can monitor workflow status:
- In the GitHub repository under "Actions" tab
- In pull requests (status checks)
- Via GitHub notifications

## Troubleshooting

### Common Issues

1. **Test Failures**: Check the specific service logs in the workflow
2. **Build Failures**: Ensure all dependencies are properly declared
3. **Database Issues**: Verify PostgreSQL service configuration
4. **Cache Issues**: Clear cache by updating go.sum or package-lock.json

### Debugging

- Workflow logs are available in the Actions tab
- Each step shows detailed output
- Failed jobs can be re-run from the Actions interface

## Adding New Services

When adding a new service:

1. Add the service name to the matrix in the workflow files
2. Ensure the service has a `make test` and `make build` target
3. Update this documentation
4. Test the workflow locally first

## Performance Optimization

- Matrix strategy runs services in parallel
- Caching reduces build times
- Selective workflow triggers prevent unnecessary runs
- Dependency management ensures efficient execution
