package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"aura-trade/api/alerts"
	"aura-trade/api/analyze"
	"aura-trade/api/chat"
	"aura-trade/api/login"
	models_api "aura-trade/api/models"
	"aura-trade/api/profile"
	"aura-trade/api/register"
	"aura-trade/api/screener"
	toppicks "aura-trade/api/top-picks"

	"aura-trade/pkg/db"
)

func main() {
	loadDotEnv()

	// Initialize DB
	database := db.GetDB()
	if database == nil {
		log.Fatal("failed to initialize database")
	}

	// Background alert checker
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			db.CheckAlerts()
		}
	}()

	mux := http.NewServeMux()

	// CORS wrapper
	c := func(next http.HandlerFunc) http.HandlerFunc {
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

	// Public routes
	mux.HandleFunc("/api/register", c(register.Handler))
	mux.HandleFunc("/api/login", c(login.Handler))
	mux.HandleFunc("/api/top-picks", c(toppicks.Handler))
	mux.HandleFunc("/api/screener", c(screener.Handler))
	mux.HandleFunc("/api/models", c(models_api.Handler))

	// Protected routes
	mux.HandleFunc("/api/chat", c(chat.Handler))
	mux.HandleFunc("/api/analyze", c(analyze.Handler))
	mux.HandleFunc("/api/alerts", c(alerts.Handler))
	mux.HandleFunc("/api/profile", c(profile.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Aura backend listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

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
