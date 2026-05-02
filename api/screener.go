package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// idxUniverse is the screener candidate pool
var idxUniverse = []string{
	"BBCA.JK", "BBRI.JK", "BMRI.JK", "TLKM.JK", "ASII.JK",
	"BREN.JK", "GOTO.JK", "UNVR.JK", "ICBP.JK", "HMSP.JK",
	"INDF.JK", "SMGR.JK", "ANTM.JK", "PTBA.JK", "ADRO.JK",
	"INCO.JK", "BUMI.JK", "MEDC.JK", "PGAS.JK", "KLBF.JK",
}

type ScreenerResult struct {
	Ticker        string  `json:"ticker"`
	CurrentPrice  float64 `json:"currentPrice"`
	PercentChange float64 `json:"percentChange"`
	Volume        int64   `json:"volume"`
	RSI           float64 `json:"rsi"`
	MA20          float64 `json:"ma20"`
	MA50          float64 `json:"ma50"`
	MACD          float64 `json:"macd"`
	MACDSignal    float64 `json:"macdSignal"`
	Signal        string  `json:"signal"` // "bullish" | "bearish" | "neutral"
}

func ScreenerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	minPrice := parseFloatQ(q.Get("minPrice"), 0)
	maxPrice := parseFloatQ(q.Get("maxPrice"), 1e9)
	minVol := parseInt64Q(q.Get("minVolume"), 0)
	minRSI := parseFloatQ(q.Get("minRSI"), 0)
	maxRSI := parseFloatQ(q.Get("maxRSI"), 100)
	signal := strings.ToLower(q.Get("signal"))

	results := screenTickers(minPrice, maxPrice, minVol, minRSI, maxRSI, signal)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func screenTickers(minPrice, maxPrice float64, minVol int64, minRSI, maxRSI float64, signal string) []ScreenerResult {
	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make([]ScreenerResult, 0, len(idxUniverse))

	// Semaphore to cap concurrent Yahoo Finance requests
	sem := make(chan struct{}, 4)

	for _, ticker := range idxUniverse {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			hist, err := fetchYahooHistory(t, 60)
			if err != nil || hist.CurrentPrice == 0 {
				return
			}

			ta := computeTA(hist)
			s := classifySignal(hist.CurrentPrice, ta)

			r := ScreenerResult{
				Ticker:        hist.Symbol,
				CurrentPrice:  hist.CurrentPrice,
				PercentChange: hist.ChangePercent,
				Volume:        hist.Volume,
				RSI:           ta.RSI,
				MA20:          ta.MA20,
				MA50:          ta.MA50,
				MACD:          ta.MACD,
				MACDSignal:    ta.MACDSignal,
				Signal:        s,
			}

			if r.CurrentPrice < minPrice || r.CurrentPrice > maxPrice {
				return
			}
			if r.Volume < minVol {
				return
			}
			if r.RSI < minRSI || r.RSI > maxRSI {
				return
			}
			if signal != "" && r.Signal != signal {
				return
			}

			mu.Lock()
			results = append(results, r)
			mu.Unlock()
		}(ticker)
	}

	wg.Wait()
	return results
}

func classifySignal(price float64, ta taResult) string {
	bullish := 0
	bearish := 0

	if ta.RSI < 45 {
		bullish++
	} else if ta.RSI > 60 {
		bearish++
	}

	if ta.MACD > ta.MACDSignal {
		bullish++
	} else {
		bearish++
	}

	if ta.MA20 > 0 && price > ta.MA20 {
		bullish++
	} else if ta.MA20 > 0 {
		bearish++
	}

	if bullish >= 2 {
		return "bullish"
	}
	if bearish >= 2 {
		return "bearish"
	}
	return "neutral"
}

func parseFloatQ(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}

func parseInt64Q(s string, def int64) int64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return v
}
