package auth

import (
	"sync"
	"time"
)

// RateLimiter tracks failed login attempts by IP address
type RateLimiter struct {
	attempts map[string]*loginAttempts
	mu       sync.RWMutex
}

type loginAttempts struct {
	count      int
	firstAttempt time.Time
	lastAttempt  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	limiter := &RateLimiter{
		attempts: make(map[string]*loginAttempts),
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// AllowLogin checks if a login attempt is allowed for an IP
func (rl *RateLimiter) AllowLogin(ip string) bool {
	rl.mu.RLock()
	attempts, exists := rl.attempts[ip]
	rl.mu.RUnlock()

	if !exists {
		return true
	}

	// Reset if more than 15 minutes have passed
	if time.Since(attempts.firstAttempt) > 15*time.Minute {
		rl.Reset(ip)
		return true
	}

	// Allow if less than 5 attempts
	return attempts.count < 5
}

// RecordAttempt records a failed login attempt
func (rl *RateLimiter) RecordAttempt(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	attempts, exists := rl.attempts[ip]
	if !exists {
		rl.attempts[ip] = &loginAttempts{
			count:        1,
			firstAttempt: time.Now(),
			lastAttempt:  time.Now(),
		}
		return
	}

	// Reset if more than 15 minutes have passed
	if time.Since(attempts.firstAttempt) > 15*time.Minute {
		attempts.count = 1
		attempts.firstAttempt = time.Now()
	} else {
		attempts.count++
	}

	attempts.lastAttempt = time.Now()
}

// Reset clears attempts for an IP (called on successful login)
func (rl *RateLimiter) Reset(ip string) {
	rl.mu.Lock()
	delete(rl.attempts, ip)
	rl.mu.Unlock()
}

// GetAttempts returns the number of attempts for an IP
func (rl *RateLimiter) GetAttempts(ip string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	attempts, exists := rl.attempts[ip]
	if !exists {
		return 0
	}

	return attempts.count
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, attempts := range rl.attempts {
			if time.Since(attempts.firstAttempt) > 15*time.Minute {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// DeployLimiter tracks deploy attempts by IP (5 deploys per minute)
type DeployLimiter struct {
	deploys map[string][]time.Time
	mu      sync.RWMutex
}

var (
	deployLimiter     *DeployLimiter
	deployLimiterOnce sync.Once
)

// GetDeployLimiter returns the singleton deploy limiter (thread-safe)
func GetDeployLimiter() *DeployLimiter {
	deployLimiterOnce.Do(func() {
		deployLimiter = &DeployLimiter{
			deploys: make(map[string][]time.Time),
		}
		go deployLimiter.cleanup()
	})
	return deployLimiter
}

// AllowDeploy checks if a deploy is allowed for an IP (max 5 per minute)
func (dl *DeployLimiter) AllowDeploy(ip string) bool {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// Filter to recent deploys only
	recent := []time.Time{}
	for _, t := range dl.deploys[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	dl.deploys[ip] = recent

	return len(recent) < 5
}

// RecordDeploy records a deploy for an IP
func (dl *DeployLimiter) RecordDeploy(ip string) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.deploys[ip] = append(dl.deploys[ip], time.Now())
}

// cleanup removes old entries periodically
func (dl *DeployLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		dl.mu.Lock()
		cutoff := time.Now().Add(-5 * time.Minute)
		for ip, times := range dl.deploys {
			recent := []time.Time{}
			for _, t := range times {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(dl.deploys, ip)
			} else {
				dl.deploys[ip] = recent
			}
		}
		dl.mu.Unlock()
	}
}
