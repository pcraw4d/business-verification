# Go Workspace Configuration Analysis

**Date:** January 2025  
**Issue:** Integration tests cannot run due to `go.work` workspace configuration

---

## Current Workspace Structure

The `go.work` file defines a multi-module workspace:

```go
go 1.24.0

toolchain go1.24.6

use ./pkg/cache
use ./pkg/performance
use ./pkg/monitoring
use ./pkg/security
use ./pkg/analytics
use ./pkg/api
use ./services/frontend
use ./services/api-gateway
use ./services/merchant-service
use (
    .
    ./cmd/railway-server
    ./services/risk-assessment-service
)
```

**Key Points:**
- The root module (`.`) is included in the workspace
- `test/integration` is part of the root module, not a separate module
- The workspace includes multiple sub-modules in `pkg/` and `services/`

---

## The Problem

When running tests with patterns like `./test/integration/...`, Go workspace mode expects:
1. Either the path to be a workspace module itself
2. Or the path to be relative to a workspace module

The error occurs because:
```
pattern ./test/integration/...: directory prefix test/integration does not contain modules listed in go.work or their selected dependencies
```

This suggests Go is looking for `test/integration` as a module, but it's actually part of the root module.

---

## Solutions

### Solution 1: Run Tests with Specific File Paths ✅ RECOMMENDED

Instead of using `./test/integration/...`, specify the exact test files:

```bash
go test -tags=integration -v -run TestWeeks24Integration \
  ./test/integration/weeks_2_4_integration_test.go \
  ./test/integration/database_setup.go
```

**Pros:**
- Works with current workspace configuration
- Explicit about which files to test
- No workspace changes needed

**Cons:**
- Must list all test files explicitly
- More verbose command

### Solution 2: Disable Workspace Mode for Tests

Use `GOWORK=off` to disable workspace mode when running tests:

```bash
GOWORK=off go test -tags=integration -v -run TestWeeks24Integration ./test/integration
```

**Pros:**
- Simple solution
- Works with standard Go test patterns
- No workspace changes needed

**Cons:**
- Must remember to set `GOWORK=off`
- May affect module resolution if tests depend on workspace modules

### Solution 3: Add Test Directory Pattern to Workspace

Modify `go.work` to explicitly include test directories (if needed):

```go
use (
    .
    ./cmd/railway-server
    ./services/risk-assessment-service
    ./test/integration  # Add this if test becomes a module
)
```

**Note:** This only works if `test/integration` becomes its own module with a `go.mod` file, which is not recommended.

### Solution 4: Use Existing Test Runner Pattern

Follow the pattern used by other test scripts in the project:

```bash
go test -run TestWeeks24Integration -v ./test/integration/weeks_2_4_integration_test.go
```

This matches the pattern used in:
- `test/compliance/run_integration_tests.sh`
- `test/compliance/run_integration_validation_tests.sh`

---

## Recommended Approach

**Use Solution 1 or Solution 2** depending on your preference:

### Option A: Specific File Paths (More Explicit)

```bash
export SUPABASE_URL="your-url"
export SUPABASE_SERVICE_ROLE_KEY="your-key"

go test -tags=integration -v -run TestWeeks24Integration \
  ./test/integration/weeks_2_4_integration_test.go \
  ./test/integration/database_setup.go
```

### Option B: Disable Workspace (More Standard)

```bash
export SUPABASE_URL="your-url"
export SUPABASE_SERVICE_ROLE_KEY="your-key"

GOWORK=off go test -tags=integration -v -run TestWeeks24Integration ./test/integration
```

---

## Updated Test Runner Script

The test runner script has been updated to use Solution 2 (GOWORK=off) as it's the most straightforward approach that maintains standard Go test patterns.

---

## Testing the Solutions

To verify which solution works best:

1. **Test Solution 1:**
   ```bash
   go test -tags=integration -v -run TestWeeks24Integration \
     ./test/integration/weeks_2_4_integration_test.go \
     ./test/integration/database_setup.go
   ```

2. **Test Solution 2:**
   ```bash
   GOWORK=off go test -tags=integration -v -run TestWeeks24Integration ./test/integration
   ```

Both should work, but Solution 2 is cleaner and more maintainable.

---

## Why This Happens

Go workspaces (`go.work`) are designed for multi-module projects where you want to:
- Develop multiple modules simultaneously
- Use local versions of dependencies
- Test changes across modules

However, when using workspaces:
- Test patterns like `./test/integration/...` can be ambiguous
- Go needs to know which module the test belongs to
- The workspace resolver may not recognize subdirectories of workspace modules

The `test/integration` directory is part of the root module (`.`), so it should work, but the `...` pattern can cause issues in workspace mode.

---

## Conclusion

The workspace configuration is correct for the project structure. The issue is with how test patterns are resolved in workspace mode. Using `GOWORK=off` or specific file paths resolves the issue without requiring workspace changes.

