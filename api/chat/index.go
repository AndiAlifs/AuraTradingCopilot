package chat

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"aura-trade/pkg/auth"
	"aura-trade/pkg/db"
	"aura-trade/pkg/logger"
	"aura-trade/pkg/models"
	"aura-trade/pkg/ollama"
	"aura-trade/pkg/yahoo"
)

type ChatRequest struct {
	History []ConversationTurn `json:"history"`
	Message string             `json:"message"`
	Model   string             `json:"model"`
}

type ConversationTurn struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type ChatResponse struct {
	Text     string                   `json:"text"`
	Strategy *models.StrategyCardData `json:"strategy,omitempty"`
}

const auraSystemPrompt = `You are Aura, the Lead System Architect & Quantitative Analyst for Aura Trading Copilot.
You specialize in IDX (Indonesian Stock Exchange) swing trading and the BSJP (Beli Sore Jual Pagi) strategy.

PERSONALITY:
- Hyper-logical, risk-averse, and highly structured.
- Speak in technical certainties and probabilities.
- You are an expert analyst, not a conversationalist. Be concise and authoritative.

BSJP STRATEGY RULES:
- Buy near market close (15:50 WIB) if technicals are bullish (RSI oversold/neutral, MACD crossover, price above MA20).
- Sell near market open (09:00 - 09:15 WIB) the next day for a quick 1-3% gain, or hold for a swing if momentum persists.
- Always set strict Stop Loss (usually 2-3% below entry or at the nearest MA support).

YOUR MANDATE:
- When a ticker is mentioned and market data is provided, YOU MUST perform an analysis.
- Provide a clear recommendation: BULLISH, BEARISH, or NEUTRAL.
- If the setup is Bullish or Bearish with >60% confidence, call 'generate_trade_setup' to render the strategy card.
- DO NOT ask the user for their outlook. YOU provide the outlook.
- If data is missing for a ticker, use 'get_stock_quote' to fetch it.`

func Handler(w http.ResponseWriter, r *http.Request) {
	auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		start := time.Now()
		apiKey := os.Getenv("GEMINI_API_KEY")
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		email := r.Header.Get("X-User-Email")
		database := db.GetDB()
		var currentUser db.User
		var rt, ps, el sql.NullString
		database.QueryRow("SELECT id, email, risk_tolerance, preferred_strategy, experience_level FROM users WHERE email = ?", email).
			Scan(&currentUser.ID, &currentUser.Email, &rt, &ps, &el)

		model := req.Model
		if model == "" {
			model = models.DefaultModel
		}

		// Context Enrichment: Detect ticker and pre-fetch data
		ticker := extractTicker(req.Message)
		var contextMessage string
		if ticker != "" {
			hist, err := yahoo.FetchYahooHistory(ticker, 60)
			if err == nil {
				ta := yahoo.ComputeTA(hist)
				contextMessage = fmt.Sprintf("[SYSTEM CONTEXT: Market data for %s - Price: %.2f, RSI: %.2f, MACD: %.2f, MA20: %.2f, ATR: %.2f. The user is asking about this ticker. Provide a recommendation using generate_trade_setup if appropriate.]",
					ticker, hist.CurrentPrice, ta.RSI, ta.MACD, ta.MA20, ta.ATR)
			}
		}

		// Log incoming request
		logger.Info("chat", logger.LogEntry{
			User:    email,
			Model:   model,
			Message: logger.Snippet(req.Message, 200),
			Ticker:  ticker,
		})

		contents := make([]models.GContent, 0, len(req.History)+2)
		for _, turn := range req.History {
			role := "user"
			if turn.Role == "aura" {
				role = "model"
			}
			contents = append(contents, models.GContent{Role: role, Parts: []models.GPart{{Text: turn.Text}}})
		}

		// Inject context if found
		if contextMessage != "" {
			contents = append(contents, models.GContent{Role: "user", Parts: []models.GPart{{Text: contextMessage}}})
		}

		contents = append(contents, models.GContent{Role: "user", Parts: []models.GPart{{Text: req.Message}}})

		text, strategy, err := runAuraAgent(apiKey, model, auraSystemPrompt, contents)
		if err != nil {
			logger.Error("chat", logger.LogEntry{
				User:       email,
				Model:      model,
				Error:      err.Error(),
				DurationMs: time.Since(start).Milliseconds(),
			})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log successful response
		logger.Info("chat", logger.LogEntry{
			User:            email,
			Model:           model,
			ResponseSnippet: logger.Snippet(text, 200),
			DurationMs:      time.Since(start).Milliseconds(),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Text: text, Strategy: strategy})
	}).ServeHTTP(w, r)
}

func extractTicker(msg string) string {
	msg = strings.ToUpper(msg)
	words := strings.Fields(msg)
	for _, w := range words {
		clean := strings.Trim(w, "?.!,()\"'")
		if strings.HasSuffix(clean, ".JK") {
			return clean
		}
		// IDX tickers are usually 4 characters
		if len(clean) >= 4 && len(clean) <= 6 {
			// Check if it's likely a ticker (all letters/numbers)
			isTicker := true
			for _, r := range clean {
				if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
					isTicker = false
					break
				}
			}
			if isTicker {
				return clean + ".JK"
			}
		}
	}
	return ""
}

func runAuraAgent(apiKey, model, systemPrompt string, contents []models.GContent) (string, *models.StrategyCardData, error) {
	if models.ProviderOf(model) == "ollama" {
		text, err := ollama.CallOllamaChat(model, systemPrompt, contents)
		return text, nil, err
	}

	var strategy *models.StrategyCardData
	for i := 0; i < 8; i++ {
		resp, err := callGeminiChat(apiKey, model, systemPrompt, contents)
		if err != nil {
			return "", nil, err
		}
		if len(resp.Candidates) == 0 {
			if resp.Error != nil {
				return "", nil, fmt.Errorf("gemini error: %s", resp.Error.Message)
			}
			return "", nil, fmt.Errorf("empty response from gemini")
		}

		candidate := resp.Candidates[0]
		var textParts []string
		var funcCalls []models.GFuncCall

		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
			if part.FunctionCall != nil {
				funcCalls = append(funcCalls, *part.FunctionCall)
			}
		}

		if len(funcCalls) == 0 {
			return strings.Join(textParts, " "), strategy, nil
		}

		contents = append(contents, models.GContent{Role: "model", Parts: candidate.Content.Parts})
		var responseParts []models.GPart
		for _, fc := range funcCalls {
			result, setupData := executeTool(fc)
			if setupData != nil {
				strategy = setupData
			}
			responseParts = append(responseParts, models.GPart{FunctionResponse: &models.GFuncResp{Name: fc.Name, Response: result}})
		}
		contents = append(contents, models.GContent{Role: "user", Parts: responseParts})
	}
	return "", nil, fmt.Errorf("agent iteration limit reached")
}

func executeTool(fc models.GFuncCall) (map[string]interface{}, *models.StrategyCardData) {
	switch fc.Name {
	case "get_stock_quote":
		ticker, _ := fc.Args["ticker"].(string)
		if ticker == "" {
			return map[string]interface{}{"error": "missing ticker"}, nil
		}
		if !strings.HasSuffix(strings.ToUpper(ticker), ".JK") {
			ticker = strings.ToUpper(ticker) + ".JK"
		}
		quote, err := yahoo.FetchYahooQuote(ticker)
		if err != nil {
			return map[string]interface{}{"error": err.Error()}, nil
		}
		return quote, nil
	case "generate_trade_setup":
		ticker, _ := fc.Args["ticker"].(string)
		setup := &models.StrategyCardData{
			Ticker:     ticker,
			Confidence: int(argFloat(fc.Args, "confidence")),
			Entry:      argFloat(fc.Args, "entry"),
			TakeProfit: argFloat(fc.Args, "takeProfit"),
			StopLoss:   argFloat(fc.Args, "stopLoss"),
			Rationale:  argString(fc.Args, "rationale"),
		}
		// Enrich with technicals if we can find them in the rationale or history
		// (Actually, since we pre-fetch, the AI already knows them)
		return map[string]interface{}{"status": "success", "message": "strategy card rendered"}, setup
	}
	return nil, nil
}

func argString(args map[string]interface{}, key string) string { v, _ := args[key].(string); return v }
func argFloat(args map[string]interface{}, key string) float64 {
	switch v := args[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case float32:
		return float64(v)
	}
	return 0
}

func callGeminiChat(apiKey, model, systemPrompt string, contents []models.GContent) (*models.GAPIResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	payload := map[string]interface{}{
		"system_instruction": map[string]interface{}{"parts": []map[string]interface{}{{"text": systemPrompt}}},
		"contents":           contents,
		"tools": []map[string]interface{}{
			{
				"function_declarations": []map[string]interface{}{
					{
						"name":        "get_stock_quote",
						"description": "Get the current stock quote for a given ticker",
						"parameters": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"ticker": map[string]interface{}{"type": "STRING", "description": "IDX ticker, e.g., BBRI.JK"},
							},
							"required": []string{"ticker"},
						},
					},
					{
						"name":        "generate_trade_setup",
						"description": "Generate a trade setup and render it as a strategy card in the UI",
						"parameters": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"ticker":     map[string]interface{}{"type": "STRING"},
								"confidence": map[string]interface{}{"type": "INTEGER", "description": "0-100 confidence score"},
								"entry":      map[string]interface{}{"type": "NUMBER"},
								"takeProfit": map[string]interface{}{"type": "NUMBER"},
								"stopLoss":   map[string]interface{}{"type": "NUMBER"},
								"rationale":  map[string]interface{}{"type": "STRING", "description": "2-sentence technical rationale"},
							},
							"required": []string{"ticker", "confidence", "entry", "takeProfit", "stopLoss", "rationale"},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResp models.GAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	return &apiResp, nil
}

