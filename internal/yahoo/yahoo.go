package yahoo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// YahooHistoryData holds extended OHLCV data for TA calculations.
type YahooHistoryData struct {
	Symbol        string
	CurrentPrice  float64
	PrevClose     float64
	Volume        int64
	FiftyTwoWeekHigh float64
	FiftyTwoWeekLow  float64
	ChangePercent float64
	Closes        []float64
	Highs         []float64
	Lows          []float64
	Volumes       []int64
}

// FetchYahooQuote returns a 5-day quote suitable for Gemini tool responses.
func FetchYahooQuote(ticker string) (map[string]interface{}, error) {
	data, err := FetchYahooHistory(ticker, 5)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"ticker":        data.Symbol,
		"currentPrice":  data.CurrentPrice,
		"previousClose": data.PrevClose,
		"changePercent": fmt.Sprintf("%.2f%%", data.ChangePercent),
		"volume":        data.Volume,
		"52weekHigh":    data.FiftyTwoWeekHigh,
		"52weekLow":     data.FiftyTwoWeekLow,
		"recentCloses":  data.Closes,
	}, nil
}

// FetchYahooHistory fetches OHLCV data for the given ticker over the specified day range.
func FetchYahooHistory(ticker string, days int) (*YahooHistoryData, error) {
	rangeParam := fmt.Sprintf("%dd", days)
	if days > 60 {
		rangeParam = "3mo"
	} else if days > 30 {
		rangeParam = "2mo"
	} else if days > 10 {
		rangeParam = "1mo"
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=%s",
		ticker, rangeParam,
	)
	log.Printf("[yahoo] → GET %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[yahoo] ✗ network error: %v", err)
		return nil, fmt.Errorf("could not reach Yahoo Finance: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("[yahoo] ← %d %s", resp.StatusCode, ticker)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var yr struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Symbol              string  `json:"symbol"`
					RegularMarketPrice  float64 `json:"regularMarketPrice"`
					ChartPreviousClose  float64 `json:"chartPreviousClose"`
					RegularMarketVolume int64   `json:"regularMarketVolume"`
					FiftyTwoWeekHigh    float64 `json:"fiftyTwoWeekHigh"`
					FiftyTwoWeekLow     float64 `json:"fiftyTwoWeekLow"`
				} `json:"meta"`
				Indicators struct {
					Quote []struct {
						Close  []float64 `json:"close"`
						High   []float64 `json:"high"`
						Low    []float64 `json:"low"`
						Volume []int64   `json:"volume"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error interface{} `json:"error"`
		} `json:"chart"`
	}

	if err := json.Unmarshal(body, &yr); err != nil {
		return nil, fmt.Errorf("failed to parse Yahoo Finance response")
	}
	if len(yr.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found for %s — check the ticker symbol", ticker)
	}

	meta := yr.Chart.Result[0].Meta
	changePercent := 0.0
	if meta.ChartPreviousClose > 0 {
		changePercent = ((meta.RegularMarketPrice - meta.ChartPreviousClose) / meta.ChartPreviousClose) * 100
	}

	result := &YahooHistoryData{
		Symbol:           meta.Symbol,
		CurrentPrice:     meta.RegularMarketPrice,
		PrevClose:        meta.ChartPreviousClose,
		Volume:           meta.RegularMarketVolume,
		FiftyTwoWeekHigh: meta.FiftyTwoWeekHigh,
		FiftyTwoWeekLow:  meta.FiftyTwoWeekLow,
		ChangePercent:    changePercent,
	}

	if len(yr.Chart.Result[0].Indicators.Quote) > 0 {
		q := yr.Chart.Result[0].Indicators.Quote[0]
		result.Closes = filterNonZero(q.Close)
		result.Highs = filterNonZero(q.High)
		result.Lows = filterNonZero(q.Low)
		result.Volumes = filterNonZeroInt(q.Volume)
	}

	log.Printf("[yahoo] parsed: %s price=%.2f change=%.2f%% days=%d",
		meta.Symbol, meta.RegularMarketPrice, changePercent, len(result.Closes))

	return result, nil
}

func filterNonZero(vals []float64) []float64 {
	out := make([]float64, 0, len(vals))
	for _, v := range vals {
		if v > 0 {
			out = append(out, v)
		}
	}
	return out
}

func filterNonZeroInt(vals []int64) []int64 {
	out := make([]int64, 0, len(vals))
	for _, v := range vals {
		if v > 0 {
			out = append(out, v)
		}
	}
	return out
}
