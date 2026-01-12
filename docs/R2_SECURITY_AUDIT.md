# Cloudflare R2 Security Audit for Screenshot Storage

**Audit Date**: January 11, 2026
**Auditor**: Security Review
**Scope**: Launch readiness of Cloudflare R2 as replacement for filesystem storage for game/world screenshots

---

## Executive Summary

The R2 integration has a solid foundation with proper path validation and private bucket configuration. However, there are **critical authorization gaps** that must be addressed before production use. Currently, any authenticated user can upload, delete, or list files in the entire bucket without ownership restrictions.

### Risk Rating: MEDIUM-HIGH

| Category | Status | Notes |
|----------|--------|-------|
| Authentication | ✅ PASS | All FileStoreService methods require authentication |
| Authorization | ❌ FAIL | No ownership checks on file operations |
| Path Security | ✅ PASS | Directory traversal prevention implemented |
| Bucket Access | ✅ PASS | Private bucket, presigned URLs for access |
| Content Validation | ⚠️ PARTIAL | No content-type or size restrictions |
| Encryption | ⚠️ UNKNOWN | R2 encryption at rest not verified |

---

## Detailed Findings

### 1. Authentication (PASS)

**Location**: `main.go:125-137`

FileStoreService is NOT in the `PublicMethods` list, meaning all operations require authentication:
- `PutFile` - requires auth
- `GetFile` - requires auth
- `DeleteFile` - requires auth
- `ListFiles` - requires auth

```go
// Only these are public (FileStoreService NOT included):
PublicMethods: []string{
    "/weewar.v1.WorldsService/ListWorlds",
    "/weewar.v1.WorldsService/GetWorld",
    // ... etc
}
```

**Status**: ✅ Authentication is properly enforced via gRPC interceptors.

---

### 2. Authorization (CRITICAL GAP)

**Location**: `services/r2/filestore.go` and `services/authz/authz.go`

**Finding**: No authorization checks exist for file operations. Any authenticated user can:

1. **Upload files anywhere**: No path ownership validation
2. **Delete any file**: No ownership check before deletion
3. **List all files**: No scope restrictions on listing
4. **Overwrite existing files**: No protection against overwriting others' screenshots

**Attack Scenarios**:

| Scenario | Impact | Likelihood |
|----------|--------|------------|
| User deletes another user's game screenshots | Data loss, griefing | HIGH |
| User uploads malicious content under legitimate paths | Reputation, legal issues | MEDIUM |
| User overwrites game/world screenshots | Data corruption | HIGH |
| User lists all files to enumerate game/world IDs | Information disclosure | MEDIUM |

**Missing Authorization Logic**:

```go
// MISSING in filestore.go - should be added:
func CanModifyFile(ctx context.Context, path string) error {
    userID := authz.GetUserIDFromContext(ctx)

    // Example: screenshots/{kind}/{id}/{theme}.{ext}
    // Should verify user owns the game/world
    parts := strings.Split(path, "/")
    if len(parts) >= 3 && parts[0] == "screenshots" {
        kind := parts[1]  // "game" or "world"
        id := parts[2]    // game/world ID

        // Verify ownership of game/world
        switch kind {
        case "game":
            game := gamesService.GetGame(id)
            if game.CreatorId != userID { return ErrNotOwner }
        case "world":
            world := worldsService.GetWorld(id)
            if world.CreatorId != userID { return ErrNotOwner }
        }
    }
    return nil
}
```

**Recommendation**: MUST implement path-based authorization before production.

---

### 3. Path Security (PASS)

**Location**: `services/r2/filestore.go:29-47`

The `validatePath()` function properly prevents:

```go
func validatePath(path string) error {
    if path == "" {
        return fmt.Errorf("path cannot be empty")
    }

    // Reject absolute paths
    if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
        return fmt.Errorf("absolute paths are not allowed: %s", path)
    }

    // Clean and check for directory traversal
    cleaned := filepath.Clean(path)
    if strings.HasPrefix(cleaned, "..") {
        return fmt.Errorf("path cannot escape root: %s", path)
    }

    return nil
}
```

**Protected Against**:
- ✅ Empty paths
- ✅ Absolute paths (`/etc/passwd`)
- ✅ Directory traversal (`../../../etc/passwd`)
- ✅ Path normalization attacks (`foo/./bar/../baz`)

**Status**: ✅ Path validation is robust.

---

### 4. Bucket Configuration (PASS)

**Location**: `main.go:208-219`

```go
r2Client, err := r2.NewR2Client(r2.R2Config{
    AccountID:       os.Getenv("R2_ACCOUNT_ID"),
    AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
    SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
    Bucket:          "weewar-assets",
    PublicURL:       "", // Private bucket - no public access
})
```

**Security Properties**:
- ✅ Private bucket (no public URL configured)
- ✅ Presigned URLs for read access (time-limited)
- ✅ Credentials loaded from environment (not hardcoded)

**Presigned URL Expiries** (`services/r2/filestore.go:50-67`):
- 15 minutes: Quick previews
- 1 hour: Default access
- 24 hours: Sharing

**Recommendation**: Consider whether 24-hour URLs are too long for sensitive content.

---

### 5. Content Validation (PARTIAL)

**Location**: `services/r2/filestore.go:69-106`

**Missing Validations**:

| Check | Status | Risk |
|-------|--------|------|
| Content-Type validation | ❌ Missing | Users could upload executables, scripts |
| File size limits | ❌ Missing | Storage exhaustion, cost DoS |
| File extension validation | ❌ Missing | Bypass content-type restrictions |
| Magic bytes verification | ❌ Missing | Content-type spoofing |

**Current Code**:
```go
contentType := req.File.ContentType
if contentType == "" {
    contentType = "application/octet-stream"  // Default accepts anything
}
```

**Recommendations**:
1. Allowlist content types for screenshots: `image/png`, `image/svg+xml`
2. Implement file size limits (e.g., 5MB max for screenshots)
3. Validate magic bytes match declared content type

---

### 6. Internal Service Calls

**Location**: `services/screenshots.go:146`

The screenshot indexer calls FileStoreService internally:

```go
resp, err := filestoreSvcClient.PutFile(context.Background(), &v1.PutFileRequest{
    File: &v1.File{Path: filePath, ContentType: contentType},
    Content: imageBytes,
})
```

**Issue**: Uses `context.Background()` which bypasses authentication.

**Analysis**: This is intentional for server-to-server calls, but highlights that:
1. The FileStoreService must distinguish between internal and external callers
2. Internal calls should have elevated privileges
3. External calls need stricter authorization

**Recommendation**: Consider implementing service-level authentication tokens for internal calls.

---

### 7. R2-Specific Security Features

**Not Verified in Codebase**:

| Feature | Cloudflare Capability | Status |
|---------|----------------------|--------|
| Encryption at Rest | R2 encrypts by default | ⚠️ Assumed enabled |
| Object Versioning | Available, not configured | ❌ Not enabled |
| Object Lock | Available, not configured | ❌ Not enabled |
| Lifecycle Rules | Available, not configured | ⚠️ Consider for cleanup |
| Access Logging | Available, not configured | ❌ Not enabled |

**Recommendations**:
1. Enable R2 object versioning for accidental deletion protection
2. Configure R2 access logging for audit trail
3. Set lifecycle rules to delete old screenshots after N days
4. Verify encryption at rest is enabled in Cloudflare dashboard

---

## Security Requirements for Launch

### P0 - Critical (Must Fix Before Launch) - ✅ IMPLEMENTED

1. **Implement File Authorization** ✅ DONE
   - Added ownership checks to `PutFile`, `DeleteFile`
   - Verifies user owns the game/world before allowing screenshot modifications
   - Internal calls (screenshot indexer) bypass auth via empty userID check
   - File: `services/filestore_validation.go`

2. **Restrict Content Types** ✅ DONE
   - Allowlist: `image/png`, `image/svg+xml` only
   - All other content types rejected with clear error message
   - File: `services/filestore_validation.go`

3. **Implement File Size Limits** ✅ DONE
   - Maximum 5MB per file for screenshots
   - Returns clear error on oversized uploads
   - File: `services/filestore_validation.go`

### P1 - High (Fix Within 30 Days)

4. **Restrict List Scope**
   - `ListFiles` should only return files the user owns
   - Or restrict to specific paths (e.g., user's games/worlds only)

5. **Enable R2 Access Logging**
   - Configure in Cloudflare dashboard
   - Store logs for security auditing

6. **Enable Object Versioning**
   - Protect against accidental deletion
   - Configure in Cloudflare dashboard

### P2 - Medium (Fix Within 90 Days)

7. **Add Magic Byte Validation**
   - Verify file content matches declared content type
   - Prevent content-type spoofing attacks

8. **Implement Rate Limiting for FileStore**
   - Prevent storage exhaustion attacks
   - Consider per-user quotas

9. **Add Lifecycle Rules**
   - Auto-delete orphaned screenshots (deleted games/worlds)
   - Reduce storage costs

---

## Implementation Plan

### Phase 1: Authorization (Blocking Launch)

```go
// In services/r2/filestore.go

// PutFile - add authorization
func (s *R2FileStoreService) PutFile(ctx context.Context, req *v1.PutFileRequest) (*v1.PutFileResponse, error) {
    // NEW: Authorization check
    if err := s.canModifyPath(ctx, req.File.Path); err != nil {
        return nil, err
    }

    // NEW: Content-type validation
    if !isAllowedContentType(req.File.ContentType) {
        return nil, fmt.Errorf("content type not allowed: %s", req.File.ContentType)
    }

    // NEW: Size limit
    if len(req.Content) > MaxFileSize {
        return nil, fmt.Errorf("file too large: %d bytes (max %d)", len(req.Content), MaxFileSize)
    }

    // ... existing code
}
```

### Phase 2: Path-Based Authorization Helper

```go
func (s *R2FileStoreService) canModifyPath(ctx context.Context, path string) error {
    // Screenshots paths: screenshots/{kind}/{id}/{theme}.{ext}
    parts := strings.Split(path, "/")
    if len(parts) < 3 {
        return authz.ErrForbidden
    }

    if parts[0] != "screenshots" {
        return authz.ErrForbidden  // Only screenshots allowed via API
    }

    kind := parts[1]
    resourceID := parts[2]

    // Internal calls bypass auth (context.Background has no user)
    userID := authz.GetUserIDFromContext(ctx)
    if userID == "" {
        // Could be internal service call - check for service token
        return nil // Or implement service auth
    }

    // Verify ownership
    switch kind {
    case "game":
        return s.verifyGameOwnership(ctx, userID, resourceID)
    case "world":
        return s.verifyWorldOwnership(ctx, userID, resourceID)
    default:
        return authz.ErrForbidden
    }
}
```

---

## Conclusion

The R2 integration is **NOW READY** for production use. All P0 security items have been implemented:

**Completed**:
1. ✅ Path-based authorization (ownership validation via GamesService/WorldsService)
2. ✅ Content-type restrictions (image/png, image/svg+xml only)
3. ✅ File size limits (5MB max)
4. ✅ Unit tests for validation logic

**Remaining P1/P2 items** (can be addressed post-launch):
- Restrict ListFiles scope
- Enable R2 access logging
- Enable object versioning
- Add magic byte validation
- Implement rate limiting for FileStore

---

## Appendix: File References

| File | Purpose | Lines |
|------|---------|-------|
| `services/filestore_validation.go` | **NEW** Shared validation & authorization | ~200 |
| `services/filestore_validation_test.go` | **NEW** Validation unit tests | ~250 |
| `services/r2/r2client.go` | R2 SDK wrapper | 172 |
| `services/r2/filestore.go` | R2 FileStore implementation | 228 |
| `services/fsbe/filestore.go` | Local filesystem FileStore | 274 |
| `services/screenshots.go` | Screenshot batch processor | 164 |
| `services/authz/authz.go` | Authorization helpers | 108 |
| `main.go` | Service registration | 251 |
