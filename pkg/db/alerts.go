package db

import (
	"log"
	"aura-trade/pkg/yahoo"
)

func CheckAlerts() {
	database := GetDB()
	rows, err := database.Query(`
		SELECT ua.id, ua.ticker, ua.` + "`condition`" + `, ua.target_price, u.email
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
		quote, err := yahoo.FetchYahooQuote(c.ticker)
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
			database.Exec("UPDATE user_alerts SET is_active = 0 WHERE id = ?", c.id)
		}
	}
}
