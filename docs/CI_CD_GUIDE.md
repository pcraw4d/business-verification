# CI/CD Guide

This guide provides comprehensive information about the Continuous Integration and Continuous Deployment (CI/CD) pipeline for the KYB Platform.

## Overview

The KYB Platform uses GitHub Actions for CI/CD automation. The pipeline includes linting, type checking, unit tests, E2E tests, and build verification.

## CI/CD Workflow

### Frontend CI/CD Workflow

Location: `.github/workflows/frontend-ci.yml`

The workflow runs on:
- Push to `main` or `develop` branches (when `frontend/**` files change)
- Pull requests to `main` or `develop` branches (when `frontend/**` files change)

### Workflow Jobs

#### 1. Lint Job

- **Purpose**: Check code quality and style
- **Tools**: ESLint
- **Duration**: ~2-3 minutes
- **Failure Action**: Warning (does not block deployment)

#### 2. Type Check Job

- **Purpose**: Verify TypeScript type correctness
- **Tools**: TypeScript compiler (`tsc --noEmit`)
- **Duration**: ~1-2 minutes
- **Failure Action**: Blocks deployment

#### 3. Unit Tests Job

- **Purpose**: Run unit tests with coverage
- **Tools**: Vitest
- **Duration**: ~3-5 minutes
- **Coverage**: Uploaded to Codecov and artifacts
- **Failure Action**: Blocks deployment

#### 4. E2E Tests Job

- **Purpose**: Run end-to-end tests
- **Tools**: Playwright
- **Duration**: ~5-10 minutes
- **Reports**: Uploaded as artifacts
- **Failure Action**: Blocks deployment

#### 5. Build Verification Job

- **Purpose**: Verify production build succeeds
- **Tools**: Next.js build
- **Duration**: ~3-5 minutes
- **Checks**: 
  - Environment variable verification
  - Build completion
  - Localhost reference check
- **Failure Action**: Blocks deployment

#### 6. All Tests Status Job

- **Purpose**: Aggregate test results
- **Duration**: ~1 minute
- **Failure Action**: Fails if any job failed

## Running Tests Locally

### Unit Tests

```bash
# Run all unit tests
cd frontend
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run tests with UI
npm run test:ui
```

### E2E Tests

```bash
# Run all E2E tests
cd frontend
npm run test:e2e

# Run E2E tests with UI
npm run test:e2e:ui

# Run E2E tests in headed mode
npm run test:e2e:headed
```

### Type Checking

```bash
# Run TypeScript type check
cd frontend
npx tsc --noEmit
```

### Linting

```bash
# Run ESLint
cd frontend
npm run lint
```

### Build Verification

```bash
# Verify build environment
cd frontend
npm run verify-env

# Build application
npm run build
```

## Test Coverage

### Coverage Thresholds

The project enforces minimum coverage thresholds:

- **Branches**: 70%
- **Functions**: 70%
- **Lines**: 70%
- **Statements**: 70%

### Viewing Coverage

```bash
# Generate coverage report
cd frontend
npm run test:coverage

# Open coverage report
open frontend/coverage/index.html
```

### Coverage Reports

Coverage reports are:
- Generated locally in `frontend/coverage/`
- Uploaded to Codecov in CI
- Available as GitHub Actions artifacts

## CI/CD Best Practices

### 1. Run Tests Before Pushing

Always run tests locally before pushing:

```bash
npm run test
npm run test:e2e
npm run lint
npx tsc --noEmit
```

### 2. Keep Tests Fast

- Unit tests should complete in < 5 minutes
- E2E tests should complete in < 10 minutes
- Use test parallelization where possible

### 3. Write Meaningful Tests

- Test behavior, not implementation
- Use descriptive test names
- Include edge cases and error scenarios

### 4. Maintain Test Coverage

- Aim for 70%+ coverage
- Focus on critical paths
- Don't sacrifice quality for coverage

### 5. Fix Failing Tests Immediately

- Don't merge PRs with failing tests
- Fix tests before adding new features
- Keep the main branch green

## Troubleshooting CI/CD Issues

### Tests Failing in CI but Passing Locally

1. Check environment variables
2. Verify Node.js version matches
3. Check for platform-specific issues
4. Review test isolation

### Build Failures

1. Check environment variable configuration
2. Verify `NEXT_PUBLIC_API_BASE_URL` is set
3. Check for TypeScript errors
4. Review build logs

### Coverage Issues

1. Verify coverage thresholds are met
2. Check for untested code paths
3. Review coverage report
4. Add missing tests

## Deployment

### Manual Deployment

Deployments to Railway are currently manual:

1. Push code to `main` branch
2. Railway automatically detects changes
3. Railway builds and deploys the service

### Automated Deployment (Future)

Future enhancement: Automated deployment via GitHub Actions:

```yaml
# .github/workflows/deploy.yml (optional)
name: Deploy

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      environment:
        required: true
        type: choice
        options:
          - staging
          - production

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Railway
        # Railway deployment steps
```

## Monitoring CI/CD

### GitHub Actions Dashboard

Monitor CI/CD status:
- View workflow runs in GitHub Actions tab
- Check individual job status
- Review test results and artifacts

### Notifications

Configure notifications for:
- Workflow failures
- Test failures
- Deployment status

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Vitest Documentation](https://vitest.dev/)
- [Playwright Documentation](https://playwright.dev/)
- [Next.js Deployment](https://nextjs.org/docs/deployment)

