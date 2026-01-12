package server

import (
	"net/http"
	"os"
)

// Google AdSense CSP domains (shared between dev and prod)
const (
	// Scripts: ad serving, tag management, ad service, traffic quality
	cspScriptSrc = "https://unpkg.com https://pagead2.googlesyndication.com https://www.googletagservices.com https://adservice.google.com https://*.google.com https://*.adtrafficquality.google"
	// Images: ad images, Google services, traffic quality
	cspImgSrc = "https://pagead2.googlesyndication.com https://www.google.com https://*.googleusercontent.com https://*.adtrafficquality.google"
	// Connections: ad syndication, traffic quality
	cspConnectSrc = "https://pagead2.googlesyndication.com https://*.adtrafficquality.google https://*.google.com"
	// Frames: ad delivery iframes, traffic quality
	cspFrameSrc = "https://googleads.g.doubleclick.net https://tpc.googlesyndication.com https://www.google.com https://*.adtrafficquality.google"
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
		// Dev adds 'unsafe-eval' for hot reloading and ws: for plain websockets
		var scriptExtra, connectExtra string
		if m.IsDevelopment {
			scriptExtra = "'unsafe-eval' "
			connectExtra = "ws: "
		}
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' "+scriptExtra+cspScriptSrc+"; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: blob: "+cspImgSrc+"; "+
				"font-src 'self'; "+
				"connect-src 'self' "+connectExtra+"wss: "+cspConnectSrc+"; "+
				"frame-src "+cspFrameSrc+"; "+
				"frame-ancestors 'none'")

		// Strict-Transport-Security: Force HTTPS (only in production)
		if !m.IsDevelopment {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
