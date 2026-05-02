package api

import (
	"encoding/json"
	"log"
	"net/http"

	"aura-trade/internal/db"
	"aura-trade/internal/auth"
)

type ProfileUpdateRequest struct {
	RiskTolerance     string `json:"riskTolerance"`
	PreferredStrategy string `json:"preferredStrategy"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Wrap with auth middleware
	auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		email := r.Header.Get("X-User-Email")
		var req ProfileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		database := db.GetDB()
		_, err := database.Exec(
			"UPDATE users SET risk_tolerance = ?, preferred_strategy = ? WHERE email = ?",
			req.RiskTolerance, req.PreferredStrategy, email,
		)
		if err != nil {
			log.Printf("[auth] profile update error: %v", err)
			http.Error(w, "failed to update profile", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}).ServeHTTP(w, r)
}
