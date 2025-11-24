# Handoff Instructions for v0.4.0 Implementation

## For the cheaper model, provide this instruction:

```
Read and execute koder/start.md
```

That's it! The file contains:
- Clear phase-by-phase implementation steps
- Test commands to verify each phase
- References to complete implementation guide
- Success criteria

## What's Ready

✅ **40+ unit tests** in `cmd/server/main_test.go`
✅ **15+ config tests** in `internal/config/config_test.go`
✅ **40+ integration tests** in `test_cli_refactor.sh`
✅ **Complete implementation guide** in `TEST_GUIDE_v0.4.0.md` (includes all code)
✅ **Quick start guide** in `koder/start.md`

## Files the Model Needs

The model will automatically reference:
- `koder/start.md` - Phase-by-phase instructions (starts here)
- `TEST_GUIDE_v0.4.0.md` - Complete code implementations
- `cmd/server/main_test.go` - Test expectations
- `koder/plans/06_cli-refactor-init-config.md` - Design document

## What the Model Will Do

1. Remove `auth.enabled` field from config
2. Implement 4 command functions (~200 lines)
3. Add CLI handlers (~150 lines)
4. Update routing (~50 lines)
5. Run tests to verify

**Estimated time:** 2-3 hours
**Total code:** ~400-500 lines

## Verification

The model should run these to confirm success:
```bash
go test ./cmd/server -v
go test ./internal/config -v
./test_cli_refactor.sh
```

All tests should pass when implementation is correct.

---

**Just say:** "Read and execute koder/start.md" 🚀
