package profile

import (
	"encoding/json"
	"net/http"

	"aura-trade/internal/db"
	"aura-trade/internal/auth"
)

type ProfileUpdateRequest struct {
	RiskTolerance     string `json:"riskTolerance"`
	PreferredStrategy string `json:"preferredStrategy"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		email := r.Header.Get("X-User-Email")
		var req ProfileUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)

		database := db.GetDB()
		database.Exec("UPDATE users SET risk_tolerance = ?, preferred_strategy = ? WHERE email = ?", req.RiskTolerance, req.PreferredStrategy, email)

		w.Write([]byte(`{"status":"success"}`))
	}).ServeHTTP(w, r)
}
