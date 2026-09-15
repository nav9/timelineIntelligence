package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"timeline-intelligence/backend/internal/service"
)

// AuthHandler handles authentication API endpoints.
// Handlers are thin wrappers — logic lives in AuthService.
type AuthHandler struct {
	authSvc     *service.AuthService
	sessionSvc  *service.SessionService
	sessionHours int
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(authSvc *service.AuthService, sessionSvc *service.SessionService, sessionHours int) *AuthHandler {
	return &AuthHandler{
		authSvc:     authSvc,
		sessionSvc:  sessionSvc,
		sessionHours: sessionHours,
	}
}

// --- Request/Response types ---

type registerRequest struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

// --- Handlers ---

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, and password are required"})
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "email address is already registered"})
		case errors.Is(err, service.ErrWeakPassword):
			// Extract the feedback from the error message.
			msg := strings.TrimPrefix(err.Error(), "password does not meet requirements: ")
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":    "password does not meet requirements",
				"feedback": strings.Split(msg, "; "),
			})
		case strings.Contains(err.Error(), "invalid email"),
			strings.Contains(err.Error(), "name is required"),
			strings.Contains(err.Error(), "email is required"):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			log.Error().Err(err).Msg("Registration error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registration successful",
		"user":    user.ToSafe(),
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	result, err := h.authSvc.Login(c.Request.Context(), service.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: ClientIP(c),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRateLimited):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many failed attempts, please wait before trying again"})
		default:
			// Generic message — do not reveal whether email exists or account status.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		}
		return
	}

	// Set the session cookie. MaxAge in seconds.
	SetSessionCookie(c, result.RawToken, h.sessionHours*3600)

	// Issue a fresh CSRF token cookie after login.
	c.SetCookie(csrfCookieName, "", -1, "/", "", false, false) // Clear old cookie.
	ensureCSRFCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"user":    result.User.ToSafe(),
	})
}

// Logout handles POST /api/auth/logout
// Requires authentication via RequireAuth middleware.
func (h *AuthHandler) Logout(c *gin.Context) {
	user := GetAuthUser(c)
	rawToken := GetSessionToken(c)

	if err := h.authSvc.Logout(c.Request.Context(), rawToken, user.ID, ClientIP(c)); err != nil {
		log.Error().Err(err).Msg("Logout error")
	}

	ClearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// Me handles GET /api/auth/me
// Returns the currently authenticated user. Requires authentication.
func (h *AuthHandler) Me(c *gin.Context) {
	user := GetAuthUser(c)
	c.JSON(http.StatusOK, gin.H{"user": user.ToSafe()})
}

// DeactivateAccount handles DELETE /api/auth/account
// Deactivates the account, revokes all sessions. Requires authentication.
func (h *AuthHandler) DeactivateAccount(c *gin.Context) {
	user := GetAuthUser(c)
	rawToken := GetSessionToken(c)

	if err := h.authSvc.DeactivateAccount(c.Request.Context(), user.ID, rawToken, ClientIP(c)); err != nil {
		log.Error().Err(err).Msg("Deactivation error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "account deactivation failed"})
		return
	}

	ClearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "account deactivated"})
}
