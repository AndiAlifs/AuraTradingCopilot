package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"aura-trade/internal/db"
	"aura-trade/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	database := db.GetDB()
	_, err = database.Exec(
		"INSERT INTO users (email, hashed_password) VALUES (?, ?)",
		req.Email, string(hashed),
	)
	if err != nil {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}

	var userID int64
	database.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&userID)
	database.Exec("INSERT INTO user_preferences (user_id) VALUES (?)", userID)

	token, err := auth.GenerateToken(req.Email)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Email: req.Email})
}
