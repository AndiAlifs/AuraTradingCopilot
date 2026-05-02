package register

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

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	json.NewDecoder(r.Body).Decode(&req)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	database := db.GetDB()
	_, err := database.Exec("INSERT INTO users (email, hashed_password) VALUES (?, ?)", req.Email, string(hashed))
	if err != nil {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}

	var userID int64
	database.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&userID)
	database.Exec("INSERT INTO user_preferences (user_id) VALUES (?)", userID)

	token, _ := auth.GenerateToken(req.Email)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": token, "email": req.Email})
}
