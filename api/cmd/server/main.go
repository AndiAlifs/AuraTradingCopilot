package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"aura-trade"
)

func main() {
	loadDotEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat", cors(api.ChatHandler))
	mux.HandleFunc("/api/top-picks", cors(api.TopPicksHandler))

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
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	}
}

// loadDotEnv reads ../.env relative to the api/ working directory.
// System env vars take precedence over .env values.
func loadDotEnv() {
	data, err := os.ReadFile("../.env")
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
