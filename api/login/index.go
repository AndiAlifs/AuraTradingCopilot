package login

import (
	"encoding/json"
	"net/http"
	"strings"

	"aura-trade/pkg/db"
	"aura-trade/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	json.NewDecoder(r.Body).Decode(&req)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	database := db.GetDB()
	var hashedPassword string
	err := database.QueryRow("SELECT hashed_password FROM users WHERE email = ?", req.Email).Scan(&hashedPassword)
	if err != nil {
		bcrypt.CompareHashAndPassword([]byte("$2a$10$placeholder"), []byte(req.Password))
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := auth.GenerateToken(req.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token, "email": req.Email})
}
