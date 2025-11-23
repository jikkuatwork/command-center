package auth

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowLogin(t *testing.T) {
	limiter := NewRateLimiter()

	// First attempt should be allowed
	if !limiter.AllowLogin("192.168.1.1") {
		t.Error("First login attempt should be allowed")
	}

	// Record 5 failed attempts
	for i := 0; i < 5; i++ {
		limiter.RecordAttempt("192.168.1.1")
	}

	// 6th attempt should be blocked
	if limiter.AllowLogin("192.168.1.1") {
		t.Error("6th login attempt should be blocked")
	}

	// Different IP should still be allowed
	if !limiter.AllowLogin("192.168.1.2") {
		t.Error("Different IP should be allowed")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter()

	// Record failed attempts
	for i := 0; i < 5; i++ {
		limiter.RecordAttempt("192.168.1.1")
	}

	// Should be blocked
	if limiter.AllowLogin("192.168.1.1") {
		t.Error("Should be blocked after 5 attempts")
	}

	// Reset
	limiter.Reset("192.168.1.1")

	// Should be allowed again
	if !limiter.AllowLogin("192.168.1.1") {
		t.Error("Should be allowed after reset")
	}
}

func TestRateLimiter_GetAttempts(t *testing.T) {
	limiter := NewRateLimiter()

	// No attempts yet
	if limiter.GetAttempts("192.168.1.1") != 0 {
		t.Error("Should have 0 attempts initially")
	}

	// Record some attempts
	limiter.RecordAttempt("192.168.1.1")
	limiter.RecordAttempt("192.168.1.1")
	limiter.RecordAttempt("192.168.1.1")

	if limiter.GetAttempts("192.168.1.1") != 3 {
		t.Errorf("Expected 3 attempts, got %d", limiter.GetAttempts("192.168.1.1"))
	}
}

func TestDeployLimiter_AllowDeploy(t *testing.T) {
	// Create fresh limiter for test
	limiter := &DeployLimiter{
		deploys: make(map[string][]time.Time),
	}

	// First 5 deploys should be allowed
	for i := 0; i < 5; i++ {
		if !limiter.AllowDeploy("192.168.1.1") {
			t.Errorf("Deploy %d should be allowed", i+1)
		}
		limiter.RecordDeploy("192.168.1.1")
	}

	// 6th deploy should be blocked
	if limiter.AllowDeploy("192.168.1.1") {
		t.Error("6th deploy should be blocked (rate limit)")
	}

	// Different IP should be allowed
	if !limiter.AllowDeploy("192.168.1.2") {
		t.Error("Different IP should be allowed")
	}
}

func TestDeployLimiter_SlidingWindow(t *testing.T) {
	limiter := &DeployLimiter{
		deploys: make(map[string][]time.Time),
	}

	// Add old deploys (> 1 minute ago)
	oldTime := time.Now().Add(-2 * time.Minute)
	limiter.deploys["192.168.1.1"] = []time.Time{oldTime, oldTime, oldTime, oldTime, oldTime}

	// Should be allowed because old deploys expired
	if !limiter.AllowDeploy("192.168.1.1") {
		t.Error("Should be allowed after old deploys expired")
	}
}
