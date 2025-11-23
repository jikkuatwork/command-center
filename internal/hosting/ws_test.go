package hosting

import (
	"testing"
	"time"
)

func TestHubManager(t *testing.T) {
	// Get hub for a site
	hub1 := GetHub("site1")
	if hub1 == nil {
		t.Fatal("GetHub() returned nil")
	}

	// Get same hub again
	hub1Again := GetHub("site1")
	if hub1 != hub1Again {
		t.Error("GetHub() should return same hub for same site")
	}

	// Get different hub for different site
	hub2 := GetHub("site2")
	if hub1 == hub2 {
		t.Error("GetHub() should return different hubs for different sites")
	}
}

func TestHubClientCount(t *testing.T) {
	hub := GetHub("test-count")

	// Initially should be 0
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("ClientCount() = %d, want 0", count)
	}
}

func TestHubBroadcast(t *testing.T) {
	hub := GetHub("test-broadcast")

	// Broadcast should not block even with no clients
	done := make(chan bool, 1)
	go func() {
		hub.Broadcast("test message")
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Broadcast() blocked with no clients")
	}
}

func TestSiteIsolation(t *testing.T) {
	// Each site should have its own hub
	sites := []string{"isolated1", "isolated2", "isolated3"}
	hubs := make([]*SiteHub, len(sites))

	for i, site := range sites {
		hubs[i] = GetHub(site)
		if hubs[i].siteID != site {
			t.Errorf("hub.siteID = %q, want %q", hubs[i].siteID, site)
		}
	}

	// Verify they're all different
	for i := 0; i < len(hubs); i++ {
		for j := i + 1; j < len(hubs); j++ {
			if hubs[i] == hubs[j] {
				t.Errorf("hubs[%d] == hubs[%d], should be different", i, j)
			}
		}
	}
}
