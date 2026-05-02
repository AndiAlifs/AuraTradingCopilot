package api

import (
	"encoding/json"
	"net/http"
)

type TopPick struct {
	Ticker        string  `json:"ticker"`
	CurrentPrice  float64 `json:"currentPrice"`
	PercentChange float64 `json:"percentChange"`
}

// TopPicksHandler returns hardcoded mock data for top picks
func TopPicksHandler(w http.ResponseWriter, r *http.Request) {
	picks := []TopPick{
		{Ticker: "BMRI.JK", CurrentPrice: 6500, PercentChange: 1.2},
		{Ticker: "BREN.JK", CurrentPrice: 5000, PercentChange: 0.5},
		{Ticker: "GOTO.JK", CurrentPrice: 50, PercentChange: -2.0},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picks)
}
