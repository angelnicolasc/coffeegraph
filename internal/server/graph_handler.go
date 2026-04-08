package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// GraphHandler manages the /graph.json endpoint and WebSocket /ws
// connections, broadcasting graph updates to all connected clients.
type GraphHandler struct {
	graphPath string

	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

// NewGraphHandler creates a handler for the given project's graph.json.
func NewGraphHandler(projectRoot string) *GraphHandler {
	return &GraphHandler{
		graphPath: filepath.Join(projectRoot, "graph.json"),
		clients:   make(map[*websocket.Conn]struct{}),
	}
}

// Register adds /graph.json and /ws routes to the mux.
func (h *GraphHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/graph.json", h.serveGraphJSON)
	mux.HandleFunc("/ws", h.handleWS)
}

// serveGraphJSON returns the current graph.json contents.
func (h *GraphHandler) serveGraphJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	b, err := os.ReadFile(h.graphPath)
	if err != nil {
		_, _ = w.Write([]byte(`{"agency":"—","nodes":[],"edges":[]}`))
		return
	}
	_, _ = w.Write(b)
}

// handleWS upgrades an HTTP connection to WebSocket and keeps it alive
// until the client disconnects.
func (h *GraphHandler) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.addClient(c)
	defer h.removeClient(c)

	// Block on reads until the client disconnects.
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *GraphHandler) addClient(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *GraphHandler) removeClient(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	_ = c.Close()
}

// broadcastGraph reads graph.json and sends it to all connected WebSocket
// clients. Clients that fail to receive are removed.
func (h *GraphHandler) broadcastGraph() {
	b, err := os.ReadFile(h.graphPath)
	if err != nil {
		return
	}

	h.mu.RLock()
	snapshot := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		snapshot = append(snapshot, c)
	}
	h.mu.RUnlock()

	for _, c := range snapshot {
		_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
			h.removeClient(c)
		}
	}
}

// StartWatching polls graph.json for changes and broadcasts updates.
// Uses a 500ms ticker and only broadcasts when the file's mtime changes.
func (h *GraphHandler) StartWatching(ctx context.Context) {
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		var lastMod time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				st, err := os.Stat(h.graphPath)
				if err != nil {
					continue
				}
				if st.ModTime().Equal(lastMod) && !lastMod.IsZero() {
					continue
				}
				lastMod = st.ModTime()
				h.broadcastGraph()
			}
		}
	}()
}
