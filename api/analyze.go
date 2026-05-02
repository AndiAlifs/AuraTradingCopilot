package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
)

// StrategyCardData is the shared output model for both the analyze endpoint
// and the chat agent's generate_trade_setup tool.
type StrategyCardData struct {
	Ticker     string  `json:"ticker"`
	Confidence int     `json:"confidence"`
	Entry      float64 `json:"entry"`
	TakeProfit float64 `json:"takeProfit"`
	StopLoss   float64 `json:"stopLoss"`
	Rationale  string  `json:"rationale"`

	// Technical Analysis (populated by /api/analyze, not the chat tool)
	RSI                     float64 `json:"rsi,omitempty"`
	MACD                    float64 `json:"macd,omitempty"`
	MACDSignal              float64 `json:"macdSignal,omitempty"`
	MACDHistogram           float64 `json:"macdHistogram,omitempty"`
	MA20                    float64 `json:"ma20,omitempty"`
	MA50                    float64 `json:"ma50,omitempty"`
	ATR                     float64 `json:"atr,omitempty"`
	StopLossJustification   string  `json:"stopLossJustification,omitempty"`
	TakeProfitJustification string  `json:"takeProfitJustification,omitempty"`
}

type AnalyzeRequest struct {
	Ticker string `json:"ticker"`
}

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[analyze] %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		http.Error(w, "GEMINI_API_KEY is not configured", http.StatusInternalServerError)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ticker := strings.ToUpper(req.Ticker)
	if !strings.HasSuffix(ticker, ".JK") {
		ticker += ".JK"
	}

	// Fetch 60 days of data for reliable TA calculations
	hist, err := fetchYahooHistory(ticker, 60)
	if err != nil {
		http.Error(w, "market data error: "+err.Error(), http.StatusBadGateway)
		return
	}

	ta := computeTA(hist)
	strategy, err := analyzeWithGemini(apiKey, ticker, hist, ta)
	if err != nil {
		http.Error(w, "gemini error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategy)
}

// ── Technical Analysis ────────────────────────────────────────────────────────

type taResult struct {
	RSI           float64
	MACD          float64
	MACDSignal    float64
	MACDHistogram float64
	MA20          float64
	MA50          float64
	ATR           float64
}

func computeTA(h *YahooHistoryData) taResult {
	closes := h.Closes
	highs := h.Highs
	lows := h.Lows

	ta := taResult{
		RSI:    calcRSI(closes, 14),
		MA20:   calcSMA(closes, 20),
		MA50:   calcSMA(closes, 50),
		ATR:    calcATR(highs, lows, closes, 14),
	}

	ta.MACD, ta.MACDSignal, ta.MACDHistogram = calcMACD(closes, 12, 26, 9)
	return ta
}

func calcSMA(closes []float64, period int) float64 {
	if len(closes) < period {
		period = len(closes)
	}
	if period == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range closes[len(closes)-period:] {
		sum += v
	}
	return sum / float64(period)
}

func calcEMA(closes []float64, period int) []float64 {
	if len(closes) == 0 || period == 0 {
		return nil
	}
	k := 2.0 / float64(period+1)
	ema := make([]float64, len(closes))
	ema[0] = closes[0]
	for i := 1; i < len(closes); i++ {
		ema[i] = closes[i]*k + ema[i-1]*(1-k)
	}
	return ema
}

func calcRSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 50
	}
	start := len(closes) - period - 1
	gains, losses := 0.0, 0.0
	for i := start + 1; i <= start+period; i++ {
		delta := closes[i] - closes[i-1]
		if delta > 0 {
			gains += delta
		} else {
			losses -= delta
		}
	}
	if losses == 0 {
		return 100
	}
	rs := (gains / float64(period)) / (losses / float64(period))
	return math.Round((100-(100/(1+rs)))*100) / 100
}

func calcMACD(closes []float64, fast, slow, signal int) (macdLine, signalLine, histogram float64) {
	if len(closes) < slow {
		return 0, 0, 0
	}
	emaFast := calcEMA(closes, fast)
	emaSlow := calcEMA(closes, slow)

	macdSeries := make([]float64, len(closes))
	for i := range closes {
		macdSeries[i] = emaFast[i] - emaSlow[i]
	}

	signalSeries := calcEMA(macdSeries, signal)
	last := len(closes) - 1
	macdLine = math.Round(macdSeries[last]*100) / 100
	signalLine = math.Round(signalSeries[last]*100) / 100
	histogram = math.Round((macdLine-signalLine)*100) / 100
	return
}

func calcATR(highs, lows, closes []float64, period int) float64 {
	n := len(closes)
	if n < 2 || len(highs) < n || len(lows) < n {
		return 0
	}
	trueRanges := make([]float64, n-1)
	for i := 1; i < n; i++ {
		hl := highs[i] - lows[i]
		hc := math.Abs(highs[i] - closes[i-1])
		lc := math.Abs(lows[i] - closes[i-1])
		trueRanges[i-1] = math.Max(hl, math.Max(hc, lc))
	}
	return math.Round(calcSMA(trueRanges, period)*100) / 100
}

// ── Gemini Analysis ───────────────────────────────────────────────────────────

func analyzeWithGemini(apiKey, ticker string, hist *YahooHistoryData, ta taResult) (*StrategyCardData, error) {
	price := hist.CurrentPrice

	// ATR-based stop loss: 1.5× ATR below current price
	atrStop := price - 1.5*ta.ATR
	// MA-based stop: slightly below MA20 if price is above it
	maStop := ta.MA20 * 0.98
	stopLoss := math.Max(atrStop, maStop)
	if stopLoss <= 0 || stopLoss >= price {
		stopLoss = price * 0.97
	}

	// Take profit: target 5-6% above entry for IDX swing trades
	takeProfit := price * 1.055

	stopJustification := fmt.Sprintf(
		"Stop set at %.0f (ATR(14)=%.0f, 1.5× ATR below entry; cross-checked against MA20 support at %.0f)",
		stopLoss, ta.ATR, ta.MA20,
	)
	tpJustification := fmt.Sprintf(
		"Target %.0f is ~5.5%% above entry, in line with IDX swing trade expectation",
		takeProfit,
	)

	taJSON, _ := json.Marshal(map[string]interface{}{
		"currentPrice":  price,
		"changePercent": fmt.Sprintf("%.2f%%", hist.ChangePercent),
		"volume":        hist.Volume,
		"52weekHigh":    hist.FiftyTwoWeekHigh,
		"52weekLow":     hist.FiftyTwoWeekLow,
		"rsi14":         ta.RSI,
		"macd":          ta.MACD,
		"macdSignal":    ta.MACDSignal,
		"macdHistogram": ta.MACDHistogram,
		"ma20":          ta.MA20,
		"ma50":          ta.MA50,
		"atr14":         ta.ATR,
		"suggestedStop": stopLoss,
		"suggestedTP":   takeProfit,
	})

	prompt := fmt.Sprintf(`You are Aura, a quantitative analyst for IDX swing trading. Be sharp and direct.

Analyze %s with this calculated market data:
%s

Respond ONLY with valid JSON, no markdown:
{
  "ticker": "string",
  "confidence": number (50–95, based on RSI momentum, MACD direction, price vs MAs),
  "entry": number (current price or a logical level),
  "takeProfit": number (use suggestedTP or adjust based on resistance),
  "stopLoss": number (use suggestedStop or tighten based on support),
  "rationale": "string (2 punchy sentences: setup thesis + catalyst)"
}`, ticker, string(taJSON))

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s",
		apiKey,
	)
	log.Printf("[gemini/analyze] → POST ticker=%s rsi=%.1f macd=%.2f ma20=%.0f", ticker, ta.RSI, ta.MACD, ta.MA20)

	body, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]interface{}{{"text": prompt}}},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.4,
			"maxOutputTokens": 512,
		},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("[gemini/analyze] ← %d", resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("parse error: %s", string(respBody))
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty Gemini response")
	}

	jsonText := geminiResp.Candidates[0].Content.Parts[0].Text
	jsonText = strings.TrimPrefix(strings.TrimSpace(jsonText), "```json")
	jsonText = strings.TrimPrefix(jsonText, "```")
	jsonText = strings.TrimSuffix(strings.TrimSpace(jsonText), "```")

	var strategy StrategyCardData
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonText)), &strategy); err != nil {
		log.Printf("[gemini/analyze] ✗ strategy parse error: %v | raw: %s", err, jsonText)
		return nil, fmt.Errorf("failed to parse strategy JSON: %v", err)
	}

	// Attach calculated TA fields
	strategy.RSI = ta.RSI
	strategy.MACD = ta.MACD
	strategy.MACDSignal = ta.MACDSignal
	strategy.MACDHistogram = ta.MACDHistogram
	strategy.MA20 = ta.MA20
	strategy.MA50 = ta.MA50
	strategy.ATR = ta.ATR
	strategy.StopLossJustification = stopJustification
	strategy.TakeProfitJustification = tpJustification

	log.Printf("[gemini/analyze] done: %s conf=%d entry=%.0f tp=%.0f sl=%.0f rsi=%.1f",
		strategy.Ticker, strategy.Confidence, strategy.Entry, strategy.TakeProfit, strategy.StopLoss, strategy.RSI)

	return &strategy, nil
}
