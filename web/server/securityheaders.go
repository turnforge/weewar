package server

import (
	"net/http"
	"os"
)

// SecurityHeadersMiddleware adds security headers to all responses.
// These headers help protect against common web vulnerabilities.
type SecurityHeadersMiddleware struct {
	// IsDevelopment disables some strict policies for local development
	IsDevelopment bool
}

// NewSecurityHeadersMiddleware creates a new security headers middleware.
// Set isDevelopment=true to relax some policies for local development.
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	isDev := os.Getenv("LILBATTLE_ENV") == "development" || os.Getenv("LILBATTLE_ENV") == ""
	return &SecurityHeadersMiddleware{
		IsDevelopment: isDev,
	}
}

// Wrap wraps an http.Handler with security headers.
func (m *SecurityHeadersMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Content-Type-Options: Prevents MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevents clickjacking by disallowing iframe embedding
		w.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection: Legacy XSS protection for older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Controls referrer information sent to other sites
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Restricts browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content-Security-Policy: Controls which resources can be loaded
		// In development, we allow unsafe-inline for easier debugging
		// In production, we use stricter policies
		if m.IsDevelopment {
			// Development: Allow inline scripts/styles for hot reloading
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com; "+
					"style-src 'self' 'unsafe-inline'; "+
					"img-src 'self' data: blob:; "+
					"font-src 'self'; "+
					"connect-src 'self' ws: wss:; "+
					"frame-ancestors 'none'")
		} else {
			// Production: Stricter CSP
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' https://unpkg.com; "+
					"style-src 'self' 'unsafe-inline'; "+ // inline styles needed for Tailwind
					"img-src 'self' data: blob:; "+
					"font-src 'self'; "+
					"connect-src 'self' wss:; "+
					"frame-ancestors 'none'")
		}

		// Strict-Transport-Security: Force HTTPS (only in production)
		if !m.IsDevelopment {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
