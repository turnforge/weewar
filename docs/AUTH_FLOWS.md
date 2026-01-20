# Authentication Flows in LilBattle

This document describes the authentication configuration for the LilBattle application.

For core authentication concepts, data models, user journeys, and edge cases, see the [oneauth documentation](https://github.com/panyam/oneauth/blob/main/docs/AUTH_FLOWS.md).

## Architecture

LilBattle uses the [oneauth](https://github.com/panyam/oneauth) library with goapplib integration. Authentication is configured in `web/server/auth.go`.

### Components

| Component | Location | Purpose |
|-----------|----------|---------|
| `setupAuthService` | `web/server/auth.go` | Creates auth stores and configures oneauth |
| `ProfilePage` | `web/server/ProfilePage.go` | Username, nickname, and password management |
| `LoginPage` | `web/server/LoginPage.go` | Login/signup form rendering |

### Data Stores

Using file-system stores (via goapplib AuthService):

```go
authService := goalservices.NewAuthService(storagePath)
usernameStore := oafs.NewFSUsernameStore(storagePath)
```

Default storage path: `~/dev-app-data/lilbattle/storage`
Override with: `LILBATTLE_USER_STORAGE_PATH`

## Signup Policy

LilBattle uses a flexible signup policy - username is NOT required at signup:

```go
SignupPolicy: &oa.SignupPolicy{
    RequireUsername:       false,  // Username added later via profile
    RequireEmail:          true,
    RequirePassword:       true,
    EnforceUsernameUnique: false,  // Not enforced at signup (no username)
    EnforceEmailUnique:    true,
    MinPasswordLength:     8,
}
```

Users can set their username later from the profile page.

## Endpoints

### Authentication Endpoints

| Endpoint | Method | Handler | Purpose |
|----------|--------|---------|---------|
| `/auth/login` | POST | `LocalAuth.HandleLogin` | Email/username + password login |
| `/auth/signup` | POST | `LocalAuth.HandleSignup` | Email + password registration |
| `/auth/google` | GET | `OAuth2` | Initiate Google OAuth |
| `/auth/google/callback` | GET | `OAuth2` | Google OAuth callback |
| `/auth/github` | GET | `OAuth2` | Initiate GitHub OAuth |
| `/auth/github/callback` | GET | `OAuth2` | GitHub OAuth callback |
| `/auth/twitter` | GET | `TwitterOAuth2` | Initiate Twitter/X OAuth |
| `/auth/twitter/callback` | GET | `TwitterOAuth2` | Twitter OAuth callback |
| `/auth/logout` | GET | `OneAuth.HandleLogout` | Logout and redirect |
| `/auth/verify-email` | GET | `LocalAuth.HandleVerifyEmail` | Verify email token |
| `/auth/resend-verification` | POST | Custom handler | Resend verification email |
| `/auth/forgot-password` | GET/POST | `LocalAuth.HandleForgotPassword` | Password reset request |
| `/auth/reset-password` | GET/POST | `LocalAuth.HandleResetPassword` | Password reset form |
| `/auth/change-password` | POST | Custom handler | Change existing password |
| `/auth/set-password` | POST | Custom handler | Set password (OAuth users) |
| `/auth/cli/token` | POST | `APIAuth` | CLI/API token authentication |

### Profile Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/profile` | GET | View profile page |
| `/profile` | POST `{action: "username"}` | Set/change username |
| `/profile` | POST `{action: "nickname"}` | Set/change nickname |

## Login Page

The login page (`/login`) supports both login and signup with tab switching.

### Login Tab
- Field: "Email or Username" (text input, not email type)
- Field: "Password"
- Auto-detects if input is email (contains `@`) or username

### Signup Tab
- Field: "Email" (email input type)
- Field: "Password"
- No username field (added later via profile)

### OAuth Providers
- Google (always available)
- GitHub (always available)
- Twitter/X (when `OAUTH2_TWITTER_CLIENT_ID` is configured)

### Error Display
Errors are stored as flash messages in session and displayed inline:
- `auth_error`: Error message text
- `auth_error_field`: Which field has the error (email/password)
- `auth_mode`: Which tab to show (login/signup)

## Profile Page

### Username Section
- Highlighted if username not set
- Format validation: 3-20 chars, lowercase, alphanumeric + `_-`
- Uniqueness enforced via UsernameStore

### Nickname Section
- Display name (not unique)
- Random nickname auto-generated on signup
- 2-30 characters

### Password Section
- Shows "Set Password" if `HasPassword=false` (OAuth-only users)
- Shows "Change Password" if `HasPassword=true`
- Current password only required when changing (not setting)

### Email Section
- Displays email with verification status
- "Resend Verification" button if not verified

## CLI/API Authentication

LilBattle supports CLI/API access via JWT tokens:

```go
apiAuth := &oa.APIAuth{
    JWTSecretKey:       jwtSecret,
    JWTIssuer:          "lilbattle",
    JWTAudience:        "cli",
    AccessTokenExpiry:  30 * 24 * time.Hour,  // 30 days
    RefreshTokenExpiry: 90 * 24 * time.Hour,  // 90 days
}
```

Get token: `POST /auth/cli/token` with email/password credentials.

## Environment Variables

```bash
# Storage path
LILBATTLE_USER_STORAGE_PATH=~/dev-app-data/lilbattle/storage

# Base URL for email links
LILBATTLE_BASE_URL=http://localhost:8080

# JWT secret for CLI tokens
JWT_CLI_SECRET=your-secret-here

# OAuth2 Providers
OAUTH2_GOOGLE_CLIENT_ID=xxx
OAUTH2_GOOGLE_CLIENT_SECRET=xxx
OAUTH2_GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback

OAUTH2_GITHUB_CLIENT_ID=xxx
OAUTH2_GITHUB_CLIENT_SECRET=xxx
OAUTH2_GITHUB_CALLBACK_URL=http://localhost:8080/auth/github/callback

OAUTH2_TWITTER_CLIENT_ID=xxx
OAUTH2_TWITTER_CLIENT_SECRET=xxx
OAUTH2_TWITTER_CALLBACK_URL=http://localhost:8080/auth/twitter/callback
```

## Email Sender

Currently using `ConsoleEmailSender` for development (prints emails to console).

For production, replace with a real email service in `web/server/auth.go`.

## Error Handling

LilBattle uses redirect-based error handling with flash messages:

```go
OnSignupError: func(err *oa.AuthError, w http.ResponseWriter, r *http.Request) bool {
    session.Put(r.Context(), "auth_error", err.Message)
    session.Put(r.Context(), "auth_error_field", err.Field)
    session.Put(r.Context(), "auth_mode", "signup")
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return true
},
```

## Security Notes

LilBattle relies on oneauth's security features:
- bcrypt password hashing
- Single-use tokens for email verification and password reset
- Generic error messages to prevent enumeration

Additional security via SCS session manager:
- HttpOnly cookies
- Secure cookies (in production)
- SameSite=Lax

### Recommendations for Production

1. Enable HTTPS (required for OAuth callbacks)
2. Set strong `JWT_CLI_SECRET`
3. Add rate limiting to auth endpoints
4. Configure real email sender
5. Enable email verification (`RequireEmailVerification: true`)

## Testing

Authentication flows are tested in oneauth's test suite. See `oneauth/user_journeys_test.go` for comprehensive journey tests covering:
- Multiple OAuth providers with same email
- OAuth user adding password
- Email signup then OAuth linking
- Username-based login
- Password and username changes
- Edge cases (race conditions, expired tokens, etc.)
