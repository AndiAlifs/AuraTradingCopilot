package analyze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"aura-trade/internal/logger"
	"aura-trade/internal/models"
	"aura-trade/internal/ollama"
	"aura-trade/internal/yahoo"
)

type AnalyzeRequest struct {
	Ticker string `json:"ticker"`
	Model  string `json:"model"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		logger.Error("analyze", logger.LogEntry{Error: "GEMINI_API_KEY is not configured"})
		http.Error(w, "GEMINI_API_KEY is not configured", http.StatusInternalServerError)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("analyze", logger.LogEntry{Error: "decode error: " + err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ticker := strings.ToUpper(req.Ticker)
	if !strings.HasSuffix(ticker, ".JK") {
		ticker += ".JK"
	}

	model := req.Model
	if model == "" {
		model = models.DefaultModel
	}

	// Log incoming request
	logger.Info("analyze", logger.LogEntry{
		Ticker: ticker,
		Model:  model,
	})

	hist, err := yahoo.FetchYahooHistory(ticker, 60)
	if err != nil {
		logger.Error("analyze", logger.LogEntry{
			Ticker:     ticker,
			Model:      model,
			Error:      "market data error: " + err.Error(),
			DurationMs: time.Since(start).Milliseconds(),
		})
		http.Error(w, "market data error: "+err.Error(), http.StatusBadGateway)
		return
	}

	ta := yahoo.ComputeTA(hist)
	strategy, err := analyzeWithGemini(apiKey, model, ticker, hist, ta)
	if err != nil {
		logger.Error("analyze", logger.LogEntry{
			Ticker:     ticker,
			Model:      model,
			Error:      "gemini error: " + err.Error(),
			DurationMs: time.Since(start).Milliseconds(),
		})
		http.Error(w, "gemini error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log successful response
	logger.Info("analyze", logger.LogEntry{
		Ticker:          ticker,
		Model:           model,
		ResponseSnippet: fmt.Sprintf("confidence=%d entry=%.0f tp=%.0f sl=%.0f", strategy.Confidence, strategy.Entry, strategy.TakeProfit, strategy.StopLoss),
		DurationMs:      time.Since(start).Milliseconds(),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategy)
}

func analyzeWithGemini(apiKey, model, ticker string, hist *yahoo.YahooHistoryData, ta yahoo.TAResult) (*models.StrategyCardData, error) {
	price := hist.CurrentPrice
	atrStop := price - 1.5*ta.ATR
	maStop := ta.MA20 * 0.98
	stopLoss := math.Max(atrStop, maStop)
	if stopLoss <= 0 || stopLoss >= price {
		stopLoss = price * 0.97
	}
	takeProfit := price * 1.055

	stopJustification := fmt.Sprintf("Stop set at %.0f (ATR(14)=%.0f, 1.5× ATR below entry; cross-checked against MA20 support at %.0f)", stopLoss, ta.ATR, ta.MA20)
	tpJustification := fmt.Sprintf("Target %.0f is ~5.5%% above entry, in line with IDX swing trade expectation", takeProfit)

	taJSON, _ := json.Marshal(map[string]interface{}{
		"currentPrice": price, "rsi14": ta.RSI, "macd": ta.MACD, "ma20": ta.MA20, "atr14": ta.ATR,
		"suggestedStop": stopLoss, "suggestedTP": takeProfit,
	})

	prompt := fmt.Sprintf(`You are Aura, an analyst for IDX swing trading. Respond ONLY with valid JSON:
{
  "ticker": "%s", "confidence": number, "entry": number, "takeProfit": number, "stopLoss": number, "rationale": "string"
}
Data: %s`, ticker, string(taJSON))

	if models.ProviderOf(model) == "ollama" {
		rawText, err := ollama.CallOllamaChat(model, "", []models.GContent{{Role: "user", Parts: []models.GPart{{Text: prompt}}}})
		if err != nil {
			return nil, err
		}
		var strategy models.StrategyCardData
		clean := strings.TrimPrefix(strings.TrimSpace(rawText), "```json")
		clean = strings.TrimSuffix(strings.TrimSpace(clean), "```")
		json.Unmarshal([]byte(clean), &strategy)
		attachTA(&strategy, ta, stopJustification, tpJustification)
		return &strategy, nil
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	payload := map[string]interface{}{"contents": []map[string]interface{}{{"parts": []map[string]interface{}{{"text": prompt}}}}}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var geminiResp struct {
		Candidates []struct{ Content struct{ Parts []struct{ Text string `json:"text"` } `json:"parts"` } `json:"content"` } `json:"candidates"`
	}
	json.NewDecoder(resp.Body).Decode(&geminiResp)
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	jsonText := geminiResp.Candidates[0].Content.Parts[0].Text
	jsonText = strings.TrimPrefix(strings.TrimSpace(jsonText), "```json")
	jsonText = strings.TrimSuffix(strings.TrimSpace(jsonText), "```")

	var strategy models.StrategyCardData
	json.Unmarshal([]byte(jsonText), &strategy)
	attachTA(&strategy, ta, stopJustification, tpJustification)

	return &strategy, nil
}

func attachTA(s *models.StrategyCardData, ta yahoo.TAResult, slj, tpj string) {
	s.RSI = ta.RSI
	s.MACD = ta.MACD
	s.MACDSignal = ta.MACDSignal
	s.MACDHistogram = ta.MACDHistogram
	s.MA20 = ta.MA20
	s.MA50 = ta.MA50
	s.ATR = ta.ATR
	s.StopLossJustification = slj
	s.TakeProfitJustification = tpj
}
