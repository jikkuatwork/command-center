package hosting

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// checkOrigin validates WebSocket origin against allowed patterns
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // Allow connections without Origin header (non-browser clients)
	}

	// Allow localhost for development
	if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
		return true
	}

	// Allow same-host connections
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	if strings.Contains(origin, host) {
		return true
	}

	log.Printf("[WS] Rejected origin: %s (host: %s)", origin, r.Host)
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// SiteHub manages WebSocket connections for a single site
type SiteHub struct {
	siteID     string
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	done       chan struct{}
	mu         sync.RWMutex
}

// HubManager manages hubs for all sites
type HubManager struct {
	hubs map[string]*SiteHub
	mu   sync.RWMutex
}

var hubManager = &HubManager{
	hubs: make(map[string]*SiteHub),
}

// GetHub returns the hub for a site, creating one if needed
func GetHub(siteID string) *SiteHub {
	hubManager.mu.Lock()
	defer hubManager.mu.Unlock()

	if hub, exists := hubManager.hubs[siteID]; exists {
		return hub
	}

	hub := &SiteHub{
		siteID:     siteID,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		done:       make(chan struct{}),
	}

	hubManager.hubs[siteID] = hub
	go hub.run()

	return hub
}

// RemoveHub stops and removes a hub for a site (call when site is deleted)
func RemoveHub(siteID string) {
	hubManager.mu.Lock()
	defer hubManager.mu.Unlock()

	if hub, exists := hubManager.hubs[siteID]; exists {
		hub.Stop()
		delete(hubManager.hubs, siteID)
		log.Printf("[WS:%s] Hub removed", siteID)
	}
}

// run handles the hub's event loop
func (h *SiteHub) run() {
	for {
		select {
		case <-h.done:
			// Shutdown: close all client connections
			h.mu.Lock()
			for conn := range h.clients {
				conn.Close()
				delete(h.clients, conn)
			}
			h.mu.Unlock()
			log.Printf("[WS:%s] Hub shutdown complete", h.siteID)
			return

		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			log.Printf("[WS:%s] Client connected (%d total)", h.siteID, len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("[WS:%s] Client disconnected (%d remaining)", h.siteID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop signals the hub to shutdown
func (h *SiteHub) Stop() {
	close(h.done)
}

// Broadcast sends a message to all connected clients
func (h *SiteHub) Broadcast(message string) {
	select {
	case h.broadcast <- []byte(message):
	default:
		// Channel full, drop message
		log.Printf("[WS:%s] Broadcast channel full, dropping message", h.siteID)
	}
}

// ClientCount returns the number of connected clients
func (h *SiteHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HandleWebSocket upgrades HTTP connections to WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request, siteID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS:%s] Upgrade error: %v", siteID, err)
		return
	}

	hub := GetHub(siteID)
	hub.register <- conn

	// Read loop - handle client messages and disconnection
	go func() {
		defer func() {
			hub.unregister <- conn
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			// Echo messages to all clients (broadcast)
			hub.Broadcast(string(message))
		}
	}()
}
