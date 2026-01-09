# Authentication System

WeeWar uses the [oneauth](https://github.com/panyam/oneauth) library for authentication, supporting both local (email/password) and OAuth2 login methods.

## Current Status

The authentication system is fully implemented and ready for configuration.

### Implemented Features

| Feature | Status | Description |
|---------|--------|-------------|
| Local Login | Ready | Email/password authentication |
| User Signup | Ready | Registration with email validation |
| Google OAuth | Needs Config | OAuth2 login via Google |
| GitHub OAuth | Needs Config | OAuth2 login via GitHub |
| Email Verification | Ready | Token-based email confirmation |
| Password Reset | Ready | Forgot password workflow |
| Change Password | Ready | Update password when logged in |
| Session Management | Ready | SCS-based session handling |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      OneAuth Library                         │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  UserStore   │  │IdentityStore│  │ChannelStore  │       │
│  │  (profiles)  │  │(email/phone)│  │ (providers)  │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│                          │                                   │
│  ┌──────────────┐        │        ┌──────────────────┐      │
│  │  TokenStore  │        │        │  OAuth Handlers  │      │
│  │ (verify/pwd) │        │        │ (Google, GitHub) │      │
│  └──────────────┘        │        └──────────────────┘      │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    File System Storage                       │
│           ~/dev-app-data/weewar/storage/                    │
│  users/  identities/  channels/  tokens/                    │
└─────────────────────────────────────────────────────────────┘
```

### Key Concepts

- **User**: Individual account with profile data
- **Identity**: Email or phone number (can be shared across providers)
- **Channel**: Authentication method (local, google, github)
- **Token**: Single-use verification/reset tokens

## Auth Routes

| Route | Method | Description |
|-------|--------|-------------|
| `/auth/google/` | GET | Initiate Google OAuth flow |
| `/auth/github/` | GET | Initiate GitHub OAuth flow |
| `/auth/login` | POST | Local login with email/password |
| `/auth/signup` | POST | User registration |
| `/auth/verify-email` | GET | Verify email via token |
| `/auth/forgot-password` | GET/POST | Request password reset |
| `/auth/reset-password` | GET/POST | Reset password with token |
| `/auth/change-password` | POST | Change password (authenticated) |
| `/auth/resend-verification` | POST | Resend verification email |

## Configuration

### Environment Variables

Copy `.env.example` to your configs folder and configure:

```bash
# Basic Configuration
WEEWAR_BASE_URL=http://localhost:8080
WEEWAR_USER_STORAGE_PATH=~/dev-app-data/weewar/storage

# Google OAuth2
OAUTH2_GOOGLE_CLIENT_ID=your-client-id
OAUTH2_GOOGLE_CLIENT_SECRET=your-client-secret
OAUTH2_GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback

# GitHub OAuth2
OAUTH2_GITHUB_CLIENT_ID=your-client-id
OAUTH2_GITHUB_CLIENT_SECRET=your-client-secret
OAUTH2_GITHUB_CALLBACK_URL=http://localhost:8080/auth/github/callback
```

### Setting Up Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new project or select existing
3. Go to "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth 2.0 Client IDs"
5. Choose "Web application"
6. Add authorized JavaScript origins: `http://localhost:8080`
7. Add authorized redirect URI: `http://localhost:8080/auth/google/callback`
8. Copy Client ID and Client Secret to your env file

### Setting Up GitHub OAuth

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in application details:
   - Homepage URL: `http://localhost:8080`
   - Authorization callback URL: `http://localhost:8080/auth/github/callback`
4. Click "Register application"
5. Copy Client ID and generate a Client Secret
6. Add to your env file

### Production Callback URLs

For production deployment, update callback URLs:

```bash
OAUTH2_GOOGLE_CALLBACK_URL=https://yourdomain.com/auth/google/callback
OAUTH2_GITHUB_CALLBACK_URL=https://yourdomain.com/auth/github/callback
```

Also update the authorized domains in Google Cloud Console and GitHub OAuth settings.

## Email Configuration

### Development

By default, emails are logged to console using `ConsoleEmailSender`:

```
Verification Email to: user@example.com
Link: http://localhost:8080/auth/verify-email?token=abc123...
```

### Production

For production, implement a real email sender. Modify `web/server/auth.go`:

```go
// Replace ConsoleEmailSender with your implementation
localAuth := &oa.LocalAuth{
    EmailSender: &YourEmailSender{}, // Implement oa.EmailSender interface
    // ...
}
```

The `EmailSender` interface requires:

```go
type EmailSender interface {
    SendVerificationEmail(to, verificationLink string) error
    SendPasswordResetEmail(to, resetLink string) error
}
```

## File Structure

```
services/
  authservice.go       # AuthService wrapping oneauth stores

web/server/
  auth.go              # Auth route handlers and setup
  webapp.go            # App integration (session, middleware)
  LoginPage.go         # Login page configuration

web/templates/
  LoginPage.html       # Login/signup UI template
```

## Testing Authentication

1. **Start the server**
   ```bash
   devloop  # or go run cmd/backend/main.go
   ```

2. **Navigate to login page**
   ```
   http://localhost:8080/login
   ```

3. **Test local signup**
   - Click "Sign Up" tab
   - Enter email and password
   - Submit and check console for verification email

4. **Test local login**
   - Enter registered email and password
   - Should redirect to home or callback URL

5. **Test OAuth** (after configuring credentials)
   - Click "Continue with Google" or "Continue with GitHub"
   - Complete OAuth flow
   - Should redirect back authenticated

## Middleware Usage

Protect routes with the auth middleware:

```go
// In your handler setup
auth := weewarApp.AuthMiddleware

// Require authentication (redirects to login)
protected := auth.EnsureUser(yourHandler)

// Extract user info (allows anonymous)
withUser := auth.ExtractUser(yourHandler)
```

Get logged-in user ID in handlers:

```go
func yourHandler(w http.ResponseWriter, r *http.Request) {
    userId := auth.GetLoggedInUserId(r)
    if userId == "" {
        // Not logged in
    }
}
```

## Troubleshooting

### OAuth Callback Errors

- **"redirect_uri_mismatch"**: Callback URL in env doesn't match Google/GitHub settings
- **"invalid_client"**: Client ID or secret is incorrect
- **"access_denied"**: User denied permission

### Email Verification Not Working

1. Check console output for verification link
2. Ensure `WEEWAR_BASE_URL` is correct
3. Verify TokenStore path is writable

### Session Not Persisting

1. Check browser allows cookies
2. Verify SCS session middleware is loaded before auth routes
3. Check for CORS issues if using different domains

## Security Considerations

- All passwords are hashed with bcrypt
- OAuth state parameter prevents CSRF
- Verification tokens are single-use
- Sessions are server-side (SCS)
- File-based stores are for development; use database for production
