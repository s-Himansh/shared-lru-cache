package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/s-Himansh/shared-lru-cache/cache"
)

type Handler struct {
	cache     *cache.ShardedCache[string, string]
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
}

type Event struct {
	Type      string                `json:"type"`
	Timestamp int64                 `json:"timestamp"`
	Data      json.RawMessage       `json:"data"`
	Metrics   cache.MetricsSnapshot `json:"metrics,omitempty"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CacheEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Shard int    `json:"shard"`
}

type CacheState struct {
	Entries   []CacheEntry          `json:"entries"`
	Shards    []int                 `json:"shards"`
	Capacity  int                   `json:"capacity"`
	NumShards int                   `json:"num_shards"`
	Size      int                   `json:"size"`
	Metrics   cache.MetricsSnapshot `json:"metrics"`
}

func NewHandler(c *cache.ShardedCache[string, string]) *Handler {
	h := &Handler{
		cache: c,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
	}
	go h.broadcastLoop()
	return h
}

func (h *Handler) broadcastLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		h.clientsMu.RLock()
		n := len(h.clients)
		h.clientsMu.RUnlock()
		if n == 0 {
			continue
		}
		state := h.getCacheState()
		data, _ := json.Marshal(state)
		h.broadcast(Event{
			Type:      "state",
			Timestamp: time.Now().UnixMilli(),
			Data:      data,
			Metrics:   state.Metrics,
		})
	}
}

func (h *Handler) broadcast(event Event) {
	data, _ := json.Marshal(event)
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func (h *Handler) broadcastEvent(eventType string, v any) {
	jsonData, _ := json.Marshal(v)
	h.broadcast(Event{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		Data:      jsonData,
		Metrics:   h.cache.Metrics(),
	})
}

func (h *Handler) getCacheState() CacheState {
	keys := h.cache.Keys()
	shards := h.cache.ShardInfo()
	sc := h.cache
	numShards := sc.NumShards()
	entries := make([]CacheEntry, 0, len(keys))

	for _, key := range keys {
		val, ok := sc.Get(key)
		if ok {
			idx := cache.StringHasher(key) % uint32(numShards)
			entries = append(entries, CacheEntry{Key: key, Value: val, Shard: int(idx)})
		}
	}

	return CacheState{
		Entries:   entries,
		Shards:    shards,
		Capacity:  sc.Capacity(),
		NumShards: numShards,
		Size:      sc.Len(),
		Metrics:   sc.Metrics(),
	}
}

// --- HTTP Handlers ---

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) State(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.getCacheState())
}

func (h *Handler) Keys(w http.ResponseWriter, r *http.Request) {
	keys := h.cache.Keys()
	writeJSON(w, map[string]any{"keys": keys, "count": len(keys)})
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	var kv KeyValue
	if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if kv.Key == "" {
		writeError(w, "key is required", http.StatusBadRequest)
		return
	}
	h.cache.Put(kv.Key, kv.Value)
	h.broadcastEvent("put", kv)
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	val, ok := h.cache.Get(req.Key)
	h.broadcastEvent("get", KeyValue{Key: req.Key, Value: val})
	writeJSON(w, map[string]any{"found": ok, "key": req.Key, "value": val})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	deleted := h.cache.Delete(req.Key)
	h.broadcastEvent("delete", map[string]any{"key": req.Key, "found": deleted})
	writeJSON(w, map[string]any{"deleted": deleted})
}

func (h *Handler) Clear(w http.ResponseWriter, r *http.Request) {
	h.cache.Clear()
	h.broadcastEvent("clear", nil)
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.cache.Metrics())
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade: %v", err)
		return
	}

	h.clientsMu.Lock()
	h.clients[conn] = true
	h.clientsMu.Unlock()

	state := h.getCacheState()
	data, _ := json.Marshal(state)
	conn.WriteMessage(websocket.TextMessage, data)

	go func() {
		defer func() {
			h.clientsMu.Lock()
			delete(h.clients, conn)
			h.clientsMu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
