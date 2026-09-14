package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/s-Himansh/shared-lru-cache/api"
	"github.com/s-Himansh/shared-lru-cache/cache"
)

func main() {
	capacity := 1000
	numShards := 16

	if v := os.Getenv("CACHE_CAPACITY"); v != "" {
		fmt.Sscanf(v, "%d", &capacity)
	}
	if v := os.Getenv("CACHE_SHARDS"); v != "" {
		fmt.Sscanf(v, "%d", &numShards)
	}

	c := cache.NewShardedCache[string, string](
		capacity, cache.StringHasher,
		cache.WithNumShards[string, string](numShards),
	)

	h := api.NewHandler(c)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/state", h.State)
	mux.HandleFunc("GET /api/keys", h.Keys)
	mux.HandleFunc("POST /api/put", h.Put)
	mux.HandleFunc("POST /api/get", h.Get)
	mux.HandleFunc("POST /api/delete", h.Delete)
	mux.HandleFunc("POST /api/clear", h.Clear)
	mux.HandleFunc("GET /api/metrics", h.Metrics)
	mux.HandleFunc("GET /ws", h.WebSocket)

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("LRU Cache API on :%s (capacity=%d, shards=%d)", port, capacity, numShards)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatal(err)
	}
}
