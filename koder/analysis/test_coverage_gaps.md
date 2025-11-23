# Test Coverage Gap Analysis

**Generated:** 2025-11-23
**Current Coverage:** auth 78.6%, config 45.5%, hosting 40%, middleware 26.6%, handlers 1.6%

---

## Executive Summary

The codebase has ~113 tests covering critical security and business logic. However, significant gaps remain in HTTP handlers (98.4% untested), the serverless runtime, and database operations.

**Key Findings:**
- Security-critical functions like `verifySignature`, `isInternalHost` need tests
- Pure utility functions are easy wins for coverage improvement
- Database-dependent code requires integration tests (lower priority)

---

## Current Test Coverage

| Package | Coverage | Tests | Status |
|---------|----------|-------|--------|
| `internal/auth` | **78.6%** | 30+ | Good |
| `internal/config` | **45.5%** | 25+ | Acceptable |
| `internal/hosting` | **40.0%** | 21 | Needs work |
| `internal/middleware` | **26.6%** | 8 | Needs work |
| `internal/handlers` | **1.6%** | 20 | Critical gap |
| `internal/database` | 0% | 0 | Integration test |
| `internal/notifier` | 0% | 0 | Low priority |
| `internal/audit` | 0% | 0 | Low priority |

---

## Untested Functions Analysis

### Package: `handlers` (98.4% untested)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `DeployHandler` | 116 | **HIGH** | High | Medium | Needs request mocking, ZIP parsing, DB |
| `LoginHandler` | 100 | **HIGH** | Medium | **Easy** | Mock session store, rate limiter |
| `StatsHandler` | 109 | Medium | Medium | Medium | Needs DB queries |
| `EventsHandler` | 71 | Medium | Medium | Medium | Pagination, DB queries |
| `TrackHandler` | 68 | **HIGH** | Low | **Easy** | Core tracking, minimal deps |
| `WebhookHandler` | 109 | Medium | Medium | Medium | HMAC verification testable |
| `PixelHandler` | 66 | Medium | Low | **Easy** | Simple GIF response |
| `RedirectHandler` | 50 | Medium | Low | **Easy** | URL lookup + redirect |
| `getClientIP` | 23 | Low | Low | **Easy** | Pure function |
| `extractDomainFromReferer` | 20 | Low | Low | **Easy** | Pure function |
| `sanitizeInput` | 15 | Low | Low | **Easy** | Pure function |
| `verifySignature` | 15 | **HIGH** | Low | **Easy** | HMAC - security critical |
| `parseInt` | 8 | Low | Low | **Easy** | Pure function |

### Package: `hosting/runtime.go` (60% untested)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `RunServerless` | 290 | **CRITICAL** | **Very High** | Hard | Goja VM, complex setup |
| `isInternalHost` | 30 | **HIGH** | Low | **Easy** | Pure SSRF check |
| `jsResponse.*` | 40 | Medium | Low | Medium | Struct methods |

### Package: `hosting/deploy.go` (partial coverage)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `ValidateAPIKey` | 27 | **HIGH** | Medium | Medium | Needs DB mock |
| `CreateAPIKey` | 26 | **HIGH** | Medium | Medium | Needs DB mock |
| `RecordDeployment` | 8 | Low | Low | Medium | DB write |
| `ListAPIKeys` | 33 | Low | Medium | Medium | DB read |
| `generateRandomToken` | 9 | Medium | Low | **Easy** | Pure crypto |

### Package: `database/db.go` (0% coverage)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `Init` | 36 | Medium | Medium | Hard | SQLite init, file system |
| `runMigrations` | 60 | Medium | High | Hard | SQL execution |
| `HealthCheck` | 7 | Low | Low | Medium | DB ping |
| `Backup` | 36 | Low | Medium | Hard | File operations |
| `cleanupOldBackups` | 47 | Low | Medium | Hard | File system |

### Package: `notifier/ntfy.go` (0% coverage)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `Send` | 50 | Low | Medium | Medium | HTTP client, mock needed |
| `CheckTrafficSpike` | 40 | Low | Medium | Hard | DB queries |
| `CheckNewDomain` | 35 | Low | Medium | Hard | DB queries |

### Package: `audit/audit.go` (0% coverage)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `LogSuccess/LogFailure` | 20 | Low | Low | Medium | DB writes |
| `GetRecentLogs` | 40 | Low | Medium | Medium | DB reads |

### Package: `auth` (21.4% untested)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `SetSessionCookie` | 15 | Medium | Low | **Easy** | HTTP cookie |
| `GetSessionCookie` | 12 | Medium | Low | **Easy** | HTTP cookie |
| `ClearSessionCookie` | 8 | Low | Low | **Easy** | HTTP cookie |

### Package: `middleware` (73.4% untested)

| Function | Lines | Criticality | Complexity | Feasibility | Notes |
|----------|-------|-------------|------------|-------------|-------|
| `SecurityHeaders` | 28 | Medium | Low | Medium | Needs config mock |

---

## Priority Matrix

### High Priority (Security Critical + Easy to Test)

| Function | Package | Why Critical | Est. Time |
|----------|---------|--------------|-----------|
| `verifySignature` | handlers/webhook | HMAC validation | 15 min |
| `isInternalHost` | hosting/runtime | SSRF protection | 15 min |
| `getClientIP` | handlers/auth | Rate limiting accuracy | 10 min |
| `sanitizeInput` | handlers/track | Input sanitization | 10 min |
| `generateRandomToken` | hosting/deploy | Token security | 10 min |
| Cookie functions | auth/cookie | Session handling | 20 min |

**Estimated total: ~1.5 hours for +15-20% coverage improvement**

### Medium Priority (Business Logic)

| Function | Package | Notes | Est. Time |
|----------|---------|-------|-----------|
| `TrackHandler` | handlers | Core tracking | 30 min |
| `PixelHandler` | handlers | High traffic endpoint | 20 min |
| `RedirectHandler` | handlers | URL handling | 20 min |
| `LoginHandler` | handlers | Auth flow | 45 min |

### Low Priority (Complex/Integration)

| Function | Package | Notes |
|----------|---------|-------|
| `RunServerless` | hosting | Requires Goja integration tests |
| Database functions | database | SQLite integration tests |
| `DeployHandler` | handlers | Full request/response cycle |

---

## Recommended Implementation Order

### Phase 1: Quick Wins (Security Critical)
```
1. isInternalHost          - SSRF protection
2. verifySignature         - Webhook HMAC
3. getClientIP             - IP extraction
4. sanitizeInput           - Input cleaning
5. generateRandomToken     - Crypto tokens
6. Cookie functions        - Session security
```

### Phase 2: Handler Utilities
```
7. extractDomainFromReferer
8. parseInt
9. checkOrigin (WebSocket)
```

### Phase 3: Core Handlers (with mocking)
```
10. TrackHandler
11. PixelHandler
12. RedirectHandler
```

### Phase 4: Integration Tests (future)
```
- Database operations
- Full deploy flow
- Serverless runtime
```

---

## Already Tested Functions (Reference)

```
TestAPIKeyOperations           TestHashPassword
TestBodySizeLimit              TestHubBroadcast
TestConfigEnvironmentMethods   TestHubClientCount
TestConfigValidate             TestHubManager
TestCreateDefaultConfig        TestInitAndGetSitesDir
TestCreateSession              TestLegacyHelpers
TestDangerousEnvVars           TestLoadFromFile
TestDefaultLimits              TestMultipleSessions
TestDeleteSession              TestNewSessionStore
TestDeployLimiter_AllowDeploy  TestPasswordRoundTrip
TestDeployLimiter_SlidingWindow TestRateLimiter_*
TestDeploySite                 TestRefreshSession
TestExpandPath                 TestRequestTracing
TestGenerateRequestID          TestSanitizeInput
TestGenerateSessionID          TestSecureFileServer
TestGetSession                 TestSiteIsolation
TestGetUserSessions            TestValidateEnvVarName
TestHasServerless              TestValidateSubdomain
```

---

## Metrics Target

| Metric | Current | Target | Gap |
|--------|---------|--------|-----|
| Overall Coverage | ~35% | 60% | 25% |
| Security Functions | ~60% | 95% | 35% |
| Handler Coverage | 1.6% | 30% | 28% |
| Auth Coverage | 78.6% | 90% | 11% |

---

## Notes

- Database-dependent tests should use SQLite in-memory mode or test fixtures
- HTTP handler tests should use `httptest.NewRecorder()` pattern
- Goja runtime tests are complex; consider integration test suite
- Focus on security-critical code first, then business logic
