package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	api "aura-trade/api"
)

func main() {
	loadDotEnv()

	// Initialize DB and run migrations on startup
	db := api.GetDB()
	if db == nil {
		log.Fatal("failed to initialize database")
	}

	// Background alert checker — runs every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			api.CheckAlerts()
		}
	}()

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/api/register", cors(api.RegisterHandler))
	mux.HandleFunc("/api/login", cors(api.LoginHandler))
	mux.HandleFunc("/api/top-picks", cors(api.TopPicksHandler))
	mux.HandleFunc("/api/screener", cors(api.ScreenerHandler))

	// Protected routes (require JWT)
	mux.HandleFunc("/api/chat", cors(api.RequireAuth(api.ChatHandler)))
	mux.HandleFunc("/api/analyze", cors(api.RequireAuth(api.AnalyzeHandler)))
	mux.HandleFunc("/api/alerts", cors(api.RequireAuth(api.AlertsHandler)))
	mux.HandleFunc("/api/profile", cors(api.RequireAuth(api.UpdateProfileHandler)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Aura backend listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// loadDotEnv reads the project-root .env (two levels above this file's package).
// The server is expected to be run from the api/ directory: go run ./cmd/server/
func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
