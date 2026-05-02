package toppicks

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"aura-trade/internal/yahoo"
)

type TopPick struct {
	Ticker        string  `json:"ticker"`
	CurrentPrice  float64 `json:"currentPrice"`
	PercentChange float64 `json:"percentChange"`
	Volume        int64   `json:"volume"`
	RSI           float64 `json:"rsi"`
	Signal        string  `json:"signal"`
}

var topPicksTickers = []string{"BBCA.JK", "BBRI.JK", "BMRI.JK", "TLKM.JK", "ASII.JK", "BREN.JK", "GOTO.JK"}

var (
	topPicksCache     []TopPick
	topPicksCacheMu   sync.RWMutex
	topPicksCacheTime time.Time
	topPicksCacheTTL  = 5 * time.Minute
)

func Handler(w http.ResponseWriter, r *http.Request) {
	topPicksCacheMu.RLock()
	cached := topPicksCache
	cacheAge := time.Since(topPicksCacheTime)
	topPicksCacheMu.RUnlock()

	if cached != nil && cacheAge < topPicksCacheTTL {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	picks := fetchTopPicks()
	topPicksCacheMu.Lock()
	topPicksCache = picks
	topPicksCacheTime = time.Now()
	topPicksCacheMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picks)
}

func fetchTopPicks() []TopPick {
	var mu sync.Mutex
	var wg sync.WaitGroup
	picks := make([]TopPick, 0, len(topPicksTickers))
	sem := make(chan struct{}, 3)

	for _, ticker := range topPicksTickers {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			hist, err := yahoo.FetchYahooHistory(t, 30)
			if err != nil {
				log.Printf("[top-picks] fetch failed %s: %v", t, err)
				return
			}
			ta := yahoo.ComputeTA(hist)
			mu.Lock()
			picks = append(picks, TopPick{
				Ticker:        hist.Symbol,
				CurrentPrice:  hist.CurrentPrice,
				PercentChange: hist.ChangePercent,
				Volume:        hist.Volume,
				RSI:           ta.RSI,
				Signal:        yahoo.ClassifySignal(hist.CurrentPrice, ta),
			})
			mu.Unlock()
		}(ticker)
	}

	wg.Wait()
	return picks
}
