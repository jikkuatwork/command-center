# Quick Start - Running Tests

## TL;DR

```bash
# 1. Try to run tests (will fail - functions don't exist yet)
go test ./cmd/server -v

# 2. Read implementation guide
cat TEST_GUIDE_v0.4.0.md

# 3. Implement functions in cmd/server/main.go

# 4. Run tests again until they pass
go test ./cmd/server -v

# 5. Build and run integration tests
make build-local
./test_cli_refactor.sh
```

## Files Overview

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/server/main_test.go` | Unit tests (40+ cases) | 895 |
| `internal/config/config_test.go` | Config tests (updated) | 386 |
| `test_cli_refactor.sh` | Integration tests | 460 |
| `TEST_GUIDE_v0.4.0.md` | Implementation guide | 800+ |
| `TEST_SUITE_SUMMARY.md` | Overview & summary | 400+ |

## What Needs Implementation

4 functions in `cmd/server/main.go`:
1. `initCommand()` - ~50 lines
2. `setCredentialsCommand()` - ~40 lines
3. `setConfigCommand()` - ~45 lines
4. `statusCommand()` - ~60 lines

Plus CLI handlers (~150 lines) and config updates (~20 lines).

**Total:** ~400-500 lines of code

## Test Commands

```bash
# Run specific command tests
go test ./cmd/server -v -run TestInitCommand
go test ./cmd/server -v -run TestSetCredentials
go test ./cmd/server -v -run TestSetConfig
go test ./cmd/server -v -run TestStatus

# Run all unit tests
go test ./cmd/server -v

# Run config tests
go test ./internal/config -v

# Run integration tests
./test_cli_refactor.sh

# Check coverage
go test ./cmd/server -cover
```

## Expected First Run

```bash
$ go test ./cmd/server -v
./main_test.go:128: undefined: initCommand
./main_test.go:372: undefined: setCredentialsCommand
./main_test.go:530: undefined: setConfigCommand
./main_test.go:719: undefined: statusCommand
FAIL [build failed]
```

This is **correct** - implement the functions and tests will pass!

## Read This First

1. `TEST_GUIDE_v0.4.0.md` - Complete implementation guide with examples
2. `koder/plans/06_cli-refactor-init-config.md` - Original design document
3. `cmd/server/main_test.go` - See what tests expect

## Success Criteria

- ✅ All 40+ unit tests pass
- ✅ All 15+ config tests pass
- ✅ All 40+ integration checks pass
- ✅ Manual CLI testing works

---

**Ready to implement!** Start with reading `TEST_GUIDE_v0.4.0.md` 📚
