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

	cors := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	}

	mux.HandleFunc("GET /api/health", cors(h.Health))
	mux.HandleFunc("GET /api/state", cors(h.State))
	mux.HandleFunc("GET /api/keys", cors(h.Keys))
	mux.HandleFunc("POST /api/put", cors(h.Put))
	mux.HandleFunc("POST /api/get", cors(h.Get))
	mux.HandleFunc("POST /api/delete", cors(h.Delete))
	mux.HandleFunc("POST /api/clear", cors(h.Clear))
	mux.HandleFunc("GET /api/metrics", cors(h.Metrics))
	mux.HandleFunc("GET /ws", h.WebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("LRU Cache API on :%s (capacity=%d, shards=%d)", port, capacity, numShards)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
