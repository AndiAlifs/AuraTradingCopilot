package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type AnalyzeRequest struct {
	Ticker string `json:"ticker"`
}

type StrategyCardData struct {
	Ticker     string  `json:"ticker"`
	Confidence int     `json:"confidence"`
	Entry      float64 `json:"entry"`
	TakeProfit float64 `json:"takeProfit"`
	StopLoss   float64 `json:"stopLoss"`
	Rationale  string  `json:"rationale"`
}

// AnalyzeHandler is a simple single-shot endpoint (non-agentic).
// The primary flow now goes through /api/chat which uses tool calling.
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
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

	quote, err := fetchYahooQuote(ticker)
	if err != nil {
		http.Error(w, "market data error: "+err.Error(), http.StatusBadGateway)
		return
	}

	currentPrice, _ := quote["currentPrice"].(float64)
	strategy, err := analyzeWithGemini(apiKey, ticker, currentPrice, quote)
	if err != nil {
		http.Error(w, "gemini error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategy)
}

func analyzeWithGemini(apiKey, ticker string, price float64, quote map[string]interface{}) (*StrategyCardData, error) {
	quoteJSON, _ := json.Marshal(quote)

	prompt := fmt.Sprintf(`You are Aura, a quantitative analyst for IDX swing trading. You are sharp and direct.

Analyze %s using this real market data:
%s

Respond ONLY with a valid JSON object, no markdown, no extra text:
{
  "ticker": "string",
  "confidence": number (50-95, based on signal clarity),
  "entry": number (current price or slightly below),
  "takeProfit": number (4-7%% above entry for IDX swing),
  "stopLoss": number (2-4%% below entry, respect support),
  "rationale": "string (2 punchy sentences: what the setup is and why now)"
}`, ticker, string(quoteJSON))

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", apiKey)

	body, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]interface{}{{"text": prompt}}},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.5,
			"maxOutputTokens": 512,
		},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
		return nil, fmt.Errorf("no response from Gemini")
	}

	jsonText := geminiResp.Candidates[0].Content.Parts[0].Text
	jsonText = strings.TrimPrefix(strings.TrimSpace(jsonText), "```json")
	jsonText = strings.TrimPrefix(jsonText, "```")
	jsonText = strings.TrimSuffix(strings.TrimSpace(jsonText), "```")

	var strategy StrategyCardData
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonText)), &strategy); err != nil {
		return nil, fmt.Errorf("failed to parse strategy JSON: %v", err)
	}

	return &strategy, nil
}
