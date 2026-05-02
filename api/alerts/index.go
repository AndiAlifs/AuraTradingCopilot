package alerts

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"aura-trade/pkg/db"
)

type Alert struct {
	ID          int     `json:"id"`
	Ticker      string  `json:"ticker"`
	Condition   string  `json:"condition"`
	TargetPrice float64 `json:"targetPrice"`
	IsActive    bool    `json:"isActive"`
	CreatedAt   string  `json:"createdAt"`
}

type CreateAlertRequest struct {
	Ticker      string  `json:"ticker"`
	Condition   string  `json:"condition"`
	TargetPrice float64 `json:"targetPrice"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	email := r.Header.Get("X-User-Email")

	switch r.Method {
	case http.MethodGet:
		getAlerts(w, r, email)
	case http.MethodPost:
		createAlert(w, r, email)
	case http.MethodDelete:
		deleteAlert(w, r, email)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAlerts(w http.ResponseWriter, _ *http.Request, email string) {
	database := db.GetDB()
	rows, err := database.Query(`
		SELECT ua.id, ua.ticker, ua.`+"`condition`"+`, ua.target_price, ua.is_active, ua.created_at
		FROM user_alerts ua
		JOIN users u ON ua.user_id = u.id
		WHERE u.email = ? AND ua.is_active = 1
		ORDER BY ua.created_at DESC
	`, email)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	alerts := []Alert{}
	for rows.Next() {
		var a Alert
		var isActive int
		if err := rows.Scan(&a.ID, &a.Ticker, &a.Condition, &a.TargetPrice, &isActive, &a.CreatedAt); err != nil {
			continue
		}
		a.IsActive = isActive == 1
		alerts = append(alerts, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func createAlert(w http.ResponseWriter, r *http.Request, email string) {
	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ticker := strings.ToUpper(strings.TrimSpace(req.Ticker))
	if !strings.HasSuffix(ticker, ".JK") {
		ticker += ".JK"
	}

	database := db.GetDB()
	var userID int64
	database.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)

	_, err := database.Exec(
		"INSERT INTO user_alerts (user_id, ticker, `condition`, target_price) VALUES (?, ?, ?, ?)",
		userID, ticker, req.Condition, req.TargetPrice,
	)
	if err != nil {
		http.Error(w, "failed to create alert", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "alert created"})
}

func deleteAlert(w http.ResponseWriter, r *http.Request, email string) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	database := db.GetDB()
	database.Exec(`
		UPDATE user_alerts SET is_active = 0
		WHERE id = ? AND user_id = (SELECT id FROM users WHERE email = ?)
	`, id, email)

	w.WriteHeader(http.StatusNoContent)
}
