# Fix Go Installation Issue

## Problem

The Go installation is corrupted, showing the error:
```
package encoding/pem is not in std (/usr/local/Cellar/go/1.24.6/libexec/src/encoding/pem)
```

This indicates that Go's standard library files exist but Go cannot recognize them as part of the standard library.

## Solution: Reinstall Go via Homebrew

Since Go is installed via Homebrew (`/usr/local/Cellar/go/1.24.6/libexec`), the best fix is to reinstall it.

### Step 1: Uninstall Go

```bash
brew uninstall go
```

### Step 2: Clean Homebrew Cache (Optional but Recommended)

```bash
brew cleanup go
```

### Step 3: Reinstall Go

```bash
brew install go
```

### Step 4: Verify Installation

```bash
# Check Go version
go version

# Verify standard library is accessible
go list std | grep encoding/pem

# Test build
cd "/Users/petercrawford/New tool"
go build ./services/classification-service/cmd/main.go
```

## Alternative Solution: Install Specific Go Version

If you need a specific Go version (e.g., 1.23.x for stability):

```bash
# Uninstall current version
brew uninstall go

# Install specific version
brew install go@1.23

# Link it (if needed)
brew link go@1.23
```

## Alternative Solution: Manual Go Installation

If Homebrew continues to have issues:

1. **Download Go from official site:**
   ```bash
   # Visit https://go.dev/dl/ and download macOS installer
   # Or use curl:
   curl -L https://go.dev/dl/go1.23.5.darwin-amd64.pkg -o /tmp/go.pkg
   ```

2. **Install the package:**
   ```bash
   sudo installer -pkg /tmp/go.pkg -target /
   ```

3. **Update PATH** (if needed):
   ```bash
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
   source ~/.zshrc
   ```

4. **Verify:**
   ```bash
   go version
   go env GOROOT
   ```

## Quick Fix: Try Rebuilding Standard Library

Sometimes the issue is just with the compiled standard library. Try:

```bash
# Clean Go cache
go clean -cache -modcache -testcache

# Rebuild standard library
cd "$(go env GOROOT)/src"
sudo ./make.bash
```

**Note**: This requires sudo access and may take a few minutes.

## Verification After Fix

Once Go is reinstalled, verify everything works:

```bash
# 1. Check Go version
go version

# 2. Verify standard library packages
go list std | head -20

# 3. Test building the classification service
cd "/Users/petercrawford/New tool"
go build ./services/classification-service/cmd/main.go

# 4. Run tests
go test -v ./test/integration/comprehensive_classification_e2e_test.go -run TestComprehensiveClassificationE2E
```

## Expected Results

After fixing:
- ✅ `go version` shows a valid Go version
- ✅ `go list std` includes `encoding/pem`
- ✅ `go build` completes without errors
- ✅ Tests can run successfully

## Troubleshooting

### If Homebrew Installation Fails

```bash
# Update Homebrew first
brew update

# Try again
brew install go
```

### If PATH Issues Occur

```bash
# Check current PATH
echo $PATH

# Verify Go is in PATH
which go

# If not, add it manually
export PATH=$PATH:/usr/local/go/bin
```

### If Permission Errors Occur

```bash
# Fix permissions on Go installation directory
sudo chown -R $(whoami) /usr/local/Cellar/go
# or
sudo chown -R $(whoami) /usr/local/go
```

## Notes

- The current Go version (1.24.6) appears to be a development version. Consider using a stable release like 1.23.x.
- After reinstalling, you may need to run `go mod tidy` in your project directories to refresh module dependencies.
- If you're using multiple Go versions, consider using `g` (Go version manager) or `gvm` for easier management.

