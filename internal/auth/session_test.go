package auth

import (
	"testing"
	"time"
)

func TestNewSessionStore(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	if store == nil {
		t.Fatal("NewSessionStore() returned nil")
	}
	if store.Count() != 0 {
		t.Errorf("New store should have 0 sessions, got %d", store.Count())
	}
}

func TestCreateSession(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	sessionID, err := store.CreateSession("testuser")
	if err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	if sessionID == "" {
		t.Error("CreateSession() returned empty session ID")
	}

	if store.Count() != 1 {
		t.Errorf("Store should have 1 session, got %d", store.Count())
	}
}

func TestCreateSessionEmptyUsername(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	_, err := store.CreateSession("")
	if err == nil {
		t.Error("CreateSession() should error for empty username")
	}
}

func TestValidateSession(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	sessionID, _ := store.CreateSession("testuser")

	valid, err := store.ValidateSession(sessionID)
	if err != nil {
		t.Fatalf("ValidateSession() error: %v", err)
	}
	if !valid {
		t.Error("ValidateSession() should return true for valid session")
	}
}

func TestValidateSessionEmpty(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	_, err := store.ValidateSession("")
	if err == nil {
		t.Error("ValidateSession() should error for empty session ID")
	}
}

func TestValidateSessionNotFound(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	valid, err := store.ValidateSession("nonexistent")
	if err != nil {
		t.Fatalf("ValidateSession() unexpected error: %v", err)
	}
	if valid {
		t.Error("ValidateSession() should return false for non-existent session")
	}
}

func TestValidateSessionExpired(t *testing.T) {
	// Create store with very short TTL
	store := NewSessionStore(1 * time.Millisecond)
	defer store.Stop()

	sessionID, _ := store.CreateSession("testuser")

	// Wait for session to expire
	time.Sleep(10 * time.Millisecond)

	valid, err := store.ValidateSession(sessionID)
	if err != nil {
		t.Fatalf("ValidateSession() error: %v", err)
	}
	if valid {
		t.Error("ValidateSession() should return false for expired session")
	}
}

func TestGetSession(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	sessionID, _ := store.CreateSession("testuser")

	session, err := store.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession() error: %v", err)
	}

	if session.Username != "testuser" {
		t.Errorf("Username = %s, want testuser", session.Username)
	}
	if session.ID != sessionID {
		t.Errorf("ID = %s, want %s", session.ID, sessionID)
	}
}

func TestGetSessionEmpty(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	_, err := store.GetSession("")
	if err == nil {
		t.Error("GetSession() should error for empty session ID")
	}
}

func TestGetSessionNotFound(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	_, err := store.GetSession("nonexistent")
	if err == nil {
		t.Error("GetSession() should error for non-existent session")
	}
}

func TestDeleteSession(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	sessionID, _ := store.CreateSession("testuser")
	if store.Count() != 1 {
		t.Fatal("Session was not created")
	}

	store.DeleteSession(sessionID)

	if store.Count() != 0 {
		t.Errorf("Store should have 0 sessions after delete, got %d", store.Count())
	}

	valid, _ := store.ValidateSession(sessionID)
	if valid {
		t.Error("Deleted session should not be valid")
	}
}

func TestDeleteSessionEmpty(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	// Should not panic
	store.DeleteSession("")
}

func TestDeleteUserSessions(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	// Create multiple sessions for same user
	store.CreateSession("user1")
	store.CreateSession("user1")
	store.CreateSession("user2")

	if store.Count() != 3 {
		t.Fatalf("Expected 3 sessions, got %d", store.Count())
	}

	count := store.DeleteUserSessions("user1")
	if count != 2 {
		t.Errorf("DeleteUserSessions() = %d, want 2", count)
	}

	if store.Count() != 1 {
		t.Errorf("Store should have 1 session remaining, got %d", store.Count())
	}
}

func TestDeleteUserSessionsEmpty(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	count := store.DeleteUserSessions("")
	if count != 0 {
		t.Errorf("DeleteUserSessions(\"\") = %d, want 0", count)
	}
}

func TestGetUserSessions(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	store.CreateSession("user1")
	store.CreateSession("user1")
	store.CreateSession("user2")

	sessions := store.GetUserSessions("user1")
	if len(sessions) != 2 {
		t.Errorf("GetUserSessions(\"user1\") returned %d sessions, want 2", len(sessions))
	}

	sessions = store.GetUserSessions("user2")
	if len(sessions) != 1 {
		t.Errorf("GetUserSessions(\"user2\") returned %d sessions, want 1", len(sessions))
	}

	sessions = store.GetUserSessions("nonexistent")
	if len(sessions) != 0 {
		t.Errorf("GetUserSessions(\"nonexistent\") returned %d sessions, want 0", len(sessions))
	}
}

func TestGetUserSessionsEmpty(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	sessions := store.GetUserSessions("")
	if sessions != nil {
		t.Error("GetUserSessions(\"\") should return nil")
	}
}

func TestRefreshSession(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	sessionID, _ := store.CreateSession("testuser")

	session, _ := store.GetSession(sessionID)
	originalExpiry := session.ExpiresAt

	// Wait a bit and refresh
	time.Sleep(10 * time.Millisecond)

	err := store.RefreshSession(sessionID)
	if err != nil {
		t.Fatalf("RefreshSession() error: %v", err)
	}

	session, _ = store.GetSession(sessionID)
	if !session.ExpiresAt.After(originalExpiry) {
		t.Error("RefreshSession() should extend expiry time")
	}
}

func TestRefreshSessionEmpty(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	err := store.RefreshSession("")
	if err == nil {
		t.Error("RefreshSession() should error for empty session ID")
	}
}

func TestRefreshSessionNotFound(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	err := store.RefreshSession("nonexistent")
	if err == nil {
		t.Error("RefreshSession() should error for non-existent session")
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID() error: %v", err)
	}

	if id1 == "" {
		t.Error("generateSessionID() returned empty string")
	}

	// Should be base64 encoded, so at least 43 characters for 32 bytes
	if len(id1) < 40 {
		t.Errorf("Session ID too short: %d chars", len(id1))
	}

	// Generate another and verify uniqueness
	id2, _ := generateSessionID()
	if id1 == id2 {
		t.Error("generateSessionID() should generate unique IDs")
	}
}

func TestMultipleSessions(t *testing.T) {
	store := NewSessionStore(time.Hour)
	defer store.Stop()

	ids := make([]string, 100)
	for i := 0; i < 100; i++ {
		id, err := store.CreateSession("user")
		if err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
		ids[i] = id
	}

	if store.Count() != 100 {
		t.Errorf("Store should have 100 sessions, got %d", store.Count())
	}

	// Verify all are unique
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			t.Error("Duplicate session ID generated")
		}
		seen[id] = true
	}
}
