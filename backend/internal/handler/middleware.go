// Package handler provides Gin HTTP handlers for the API.
//
// Handlers are intentionally thin: they parse request input, call service
// methods, and format responses. Business logic belongs in the service layer.
//
// middleware.go: Authentication, CSRF, and security-header middleware.
package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"timeline-intelligence/backend/internal/domain"
	"timeline-intelligence/backend/internal/service"
)

const (
	// sessionCookieName is the name of the HttpOnly session cookie.
	sessionCookieName = "tl_session"

	// csrfCookieName is the name of the readable (non-HttpOnly) CSRF token cookie.
	csrfCookieName = "tl_csrf"

	// csrfHeaderName is the header the frontend must include on state-changing requests.
	csrfHeaderName = "X-CSRF-Token"

	// contextKeyUser is the key used to store the authenticated user in gin.Context.
	contextKeyUser = "auth_user"

	// contextKeySessionToken is used to pass the raw session token to handlers.
	contextKeySessionToken = "session_token"
)

// SecurityHeaders adds protective HTTP response headers to every response.
// These provide defense-in-depth against common browser attacks.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "0") // Disabled — modern CSP is preferred
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// Content Security Policy.
		// Allows: same-origin scripts/styles, no inline scripts (safe default).
		// Adjust if bundled assets use inline styles (Vite typically does not in prod).
		csp := strings.Join([]string{
			"default-src 'self'",
			"script-src 'self'",
			"style-src 'self' 'unsafe-inline'", // Allow inline styles for Svelte
			"img-src 'self' data:",
			"font-src 'self'",
			"connect-src 'self'",
			"frame-ancestors 'none'",
		}, "; ")
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}

// RequireAuth is middleware that verifies the session cookie.
// If the session is valid, the authenticated user is attached to the context.
// If not, a 401 response is returned and the request is aborted.
func RequireAuth(sessionSvc *service.SessionService, authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, err := c.Cookie(sessionCookieName)
		if err != nil || rawToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		session, err := sessionSvc.ValidateSession(ctx, rawToken)
		if err != nil {
			log.Error().Err(err).Msg("Session validation error")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		if session == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		user, err := authSvc.GetUserByID(ctx, session.UserID)
		if err != nil || user == nil || !user.IsActive() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		c.Set(contextKeyUser, user)
		c.Set(contextKeySessionToken, rawToken)
		c.Next()
	}
}

// RequireAdmin is middleware that allows only admin-role users through.
// Must be used after RequireAuth.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthUser(c)
		if user == nil || !user.IsAdmin() {
			c.JSON(http.StatusForbidden, gin.H{"error": "administrator access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CSRFProtect validates the CSRF token on state-changing requests.
// The CSRF token is issued as a readable (non-HttpOnly) cookie so the
// frontend JavaScript can read it and include it as a header.
// The middleware checks that the header value matches the cookie value.
//
// This is the Double-Submit Cookie pattern.
// It is effective because a cross-origin attacker cannot read cookies.
func CSRFProtect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only protect state-changing methods.
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			ensureCSRFCookie(c)
			c.Next()
			return
		}

		cookieToken, cookieErr := c.Cookie(csrfCookieName)
		headerToken := c.GetHeader(csrfHeaderName)

		if cookieErr != nil || cookieToken == "" || headerToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
			c.Abort()
			return
		}

		if cookieToken != headerToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token invalid"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ensureCSRFCookie issues a CSRF token cookie if one is not already present.
func ensureCSRFCookie(c *gin.Context) {
	if _, err := c.Cookie(csrfCookieName); err == nil {
		return // Cookie already exists.
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return
	}
	token := hex.EncodeToString(b)
	// SameSite=Strict; NOT HttpOnly (frontend must read it).
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set true in production with TLS.
	})
}

// GetAuthUser retrieves the authenticated user from the gin context.
// Returns nil if not set (should not happen after RequireAuth middleware).
func GetAuthUser(c *gin.Context) *domain.User {
	if v, exists := c.Get(contextKeyUser); exists {
		if u, ok := v.(*domain.User); ok {
			return u
		}
	}
	return nil
}

// GetSessionToken retrieves the raw session token from the gin context.
func GetSessionToken(c *gin.Context) string {
	if v, exists := c.Get(contextKeySessionToken); exists {
		if t, ok := v.(string); ok {
			return t
		}
	}
	return ""
}

// ClientIP extracts the real client IP from the request, considering X-Forwarded-For.
// In production behind a trusted reverse proxy this would need more careful handling.
func ClientIP(c *gin.Context) string {
	return c.ClientIP()
}

// SetSessionCookie sets the HttpOnly session cookie.
func SetSessionCookie(c *gin.Context, rawToken string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    rawToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set true in production with TLS.
	})
}

// ClearSessionCookie removes the session cookie.
func ClearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
