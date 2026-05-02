package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Alert struct {
	ID          int     `json:"id"`
	Ticker      string  `json:"ticker"`
	Condition   string  `json:"condition"` // "above" or "below"
	TargetPrice float64 `json:"targetPrice"`
	IsActive    bool    `json:"isActive"`
	CreatedAt   string  `json:"createdAt"`
}

type CreateAlertRequest struct {
	Ticker      string  `json:"ticker"`
	Condition   string  `json:"condition"`
	TargetPrice float64 `json:"targetPrice"`
}

// AlertsHandler routes GET/POST/DELETE for /api/alerts (requires auth).
func AlertsHandler(w http.ResponseWriter, r *http.Request) {
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
	db := GetDB()
	rows, err := db.Query(`
		SELECT ua.id, ua.ticker, ua.condition, ua.target_price, ua.is_active, ua.created_at
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
	if req.Condition != "above" && req.Condition != "below" {
		http.Error(w, "condition must be 'above' or 'below'", http.StatusBadRequest)
		return
	}
	if req.TargetPrice <= 0 {
		http.Error(w, "target price must be positive", http.StatusBadRequest)
		return
	}

	db := GetDB()
	var userID int64
	if err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	result, err := db.Exec(
		"INSERT INTO user_alerts (user_id, ticker, condition, target_price) VALUES (?, ?, ?, ?)",
		userID, ticker, req.Condition, req.TargetPrice,
	)
	if err != nil {
		http.Error(w, "failed to create alert", http.StatusInternalServerError)
		return
	}

	alertID, _ := result.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      alertID,
		"message": "alert created",
	})
}

func deleteAlert(w http.ResponseWriter, r *http.Request, email string) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid alert id", http.StatusBadRequest)
		return
	}

	db := GetDB()
	res, err := db.Exec(`
		UPDATE user_alerts SET is_active = 0
		WHERE id = ? AND user_id = (SELECT id FROM users WHERE email = ?)
	`, id, email)
	if err != nil {
		http.Error(w, "failed to delete alert", http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "alert not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckAlerts fetches active alerts and logs any that have been triggered.
// Call this periodically (e.g., every 5 minutes) from a goroutine in main.
func CheckAlerts() {
	db := GetDB()
	rows, err := db.Query(`
		SELECT ua.id, ua.ticker, ua.condition, ua.target_price, u.email
		FROM user_alerts ua
		JOIN users u ON ua.user_id = u.id
		WHERE ua.is_active = 1
	`)
	if err != nil {
		log.Printf("[alerts] error fetching active alerts: %v", err)
		return
	}
	defer rows.Close()

	type check struct {
		id          int
		ticker      string
		condition   string
		targetPrice float64
		email       string
	}

	var checks []check
	for rows.Next() {
		var c check
		if err := rows.Scan(&c.id, &c.ticker, &c.condition, &c.targetPrice, &c.email); err != nil {
			continue
		}
		checks = append(checks, c)
	}
	rows.Close()

	for _, c := range checks {
		quote, err := fetchYahooQuote(c.ticker)
		if err != nil {
			log.Printf("[alerts] quote fetch failed %s: %v", c.ticker, err)
			continue
		}
		current, _ := quote["currentPrice"].(float64)

		triggered := (c.condition == "above" && current >= c.targetPrice) ||
			(c.condition == "below" && current <= c.targetPrice)

		if triggered {
			log.Printf("[alerts] TRIGGERED %s %s %.2f (now %.2f) → %s",
				c.ticker, c.condition, c.targetPrice, current, c.email)
			db.Exec("UPDATE user_alerts SET is_active = 0 WHERE id = ?", c.id)
		}
	}
}
