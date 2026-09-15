package service

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"timeline-intelligence/backend/internal/repository"
)

// RateLimitService provides IP-based progressive login throttling.
//
// Failure counts are tracked in two places:
//   1. In-memory (fast, resets on restart)
//   2. In the auth_events database table (persisted, used on startup)
//
// Progressive delays (seconds after N failures):
//   1: 1s, 2: 2s, 3: 4s, 4: 8s, 5+: 30s
//
// The delay is applied by the auth service BEFORE returning the error
// to the caller, making it impossible to bypass via the frontend.
type RateLimitService struct {
	mu          sync.Mutex
	failures    map[string][]time.Time // IP → failure timestamps
	auditRepo   repository.AuditRepository
	maxAttempts int
	baseDelay   time.Duration
	window      time.Duration // Rolling window for counting failures
}

// NewRateLimitService creates a RateLimitService.
func NewRateLimitService(
	auditRepo repository.AuditRepository,
	maxAttempts int,
	baseDelaySeconds int,
) *RateLimitService {
	return &RateLimitService{
		failures:    make(map[string][]time.Time),
		auditRepo:   auditRepo,
		maxAttempts: maxAttempts,
		baseDelay:   time.Duration(baseDelaySeconds) * time.Second,
		window:      15 * time.Minute,
	}
}

// RecordFailure records a failed attempt for an IP address.
func (r *RateLimitService) RecordFailure(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failures[ip] = append(r.failures[ip], time.Now())
	r.pruneOld(ip)
}

// RecordSuccess clears the in-memory failure counter for an IP after a successful login.
func (r *RateLimitService) RecordSuccess(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.failures, ip)
}

// IsRateLimited returns true if the IP has exceeded the maximum failure threshold.
func (r *RateLimitService) IsRateLimited(ctx context.Context, ip string) bool {
	count := r.recentFailureCount(ctx, ip)
	return count >= r.maxAttempts
}

// DelayForIP returns the delay duration the caller must sleep before responding
// to a failed login attempt from this IP.
func (r *RateLimitService) DelayForIP(ctx context.Context, ip string) time.Duration {
	count := r.recentFailureCount(ctx, ip)
	return r.delayForCount(count)
}

// recentFailureCount returns the number of failures in the sliding window,
// consulting the in-memory store (fast). Falls back to the DB if the
// in-memory store was recently cleared (e.g. after restart).
func (r *RateLimitService) recentFailureCount(ctx context.Context, ip string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	since := time.Now().Add(-r.window)
	count := 0
	for _, t := range r.failures[ip] {
		if t.After(since) {
			count++
		}
	}

	// If in-memory count is zero, check the database (handles restart scenario).
	if count == 0 {
		dbCount, err := r.auditRepo.CountRecentFailuresByIP(ctx, ip, since)
		if err != nil {
			log.Warn().Err(err).Str("ip", ip).Msg("Failed to query DB failure count")
			return 0
		}
		return dbCount
	}
	return count
}

// pruneOld removes failure timestamps outside the current window. Must be called with mu held.
func (r *RateLimitService) pruneOld(ip string) {
	since := time.Now().Add(-r.window)
	fresh := r.failures[ip][:0]
	for _, t := range r.failures[ip] {
		if t.After(since) {
			fresh = append(fresh, t)
		}
	}
	r.failures[ip] = fresh
}

// delayForCount returns the progressive delay for a given failure count.
// Delay doubles with each additional failure, capped at 30 seconds.
func (r *RateLimitService) delayForCount(count int) time.Duration {
	if count <= 0 {
		return 0
	}
	delay := r.baseDelay
	for i := 1; i < count && delay < 30*time.Second; i++ {
		delay *= 2
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
	}
	return delay
}
