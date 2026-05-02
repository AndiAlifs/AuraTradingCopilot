package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"aura-trade/internal/db"
	"aura-trade/internal/models"
	"aura-trade/internal/ollama"
	"aura-trade/internal/yahoo"
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

const auraSystemPrompt = `You are Aura, an AI quantitative analyst built for Indonesian Stock Exchange (IDX) swing trading. Sharp, confident, trader-brained.
Tools: get_stock_quote, generate_trade_setup.
Rules: Always fetch data before setups. Target 4-7% profit, 2-4% stop. 1-3 sentences max response.`

var auraTools = []map[string]interface{}{
	{
		"functionDeclarations": []map[string]interface{}{
			{
				"name":        "get_stock_quote",
				"description": "Fetch real-time stock price from Yahoo Finance",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"ticker": map[string]interface{}{"type": "string", "description": "IDX ticker with .JK suffix"},
					},
					"required": []string{"ticker"},
				},
			},
			{
				"name":        "generate_trade_setup",
				"description": "Display trade card",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"ticker":     map[string]interface{}{"type": "string"},
						"confidence": map[string]interface{}{"type": "number"},
						"entry":      map[string]interface{}{"type": "number"},
						"takeProfit": map[string]interface{}{"type": "number"},
						"stopLoss":   map[string]interface{}{"type": "number"},
						"rationale":  map[string]interface{}{"type": "string"},
					},
					"required": []string{"ticker", "confidence", "entry", "takeProfit", "stopLoss", "rationale"},
				},
			},
		},
	},
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		http.Error(w, "GEMINI_API_KEY not configured", http.StatusInternalServerError)
		return
	}

	var req ChatRequest
	json.NewDecoder(r.Body).Decode(&req)

	email := r.Header.Get("X-User-Email")
	var currentUser db.User
	var rt, ps, el sql.NullString
	database := db.GetDB()
	database.QueryRow("SELECT id, email, risk_tolerance, preferred_strategy, experience_level FROM users WHERE email = ?", email).
		Scan(&currentUser.ID, &currentUser.Email, &rt, &ps, &el)
	currentUser.RiskTolerance = rt.String
	currentUser.PreferredStrategy = ps.String
	currentUser.ExperienceLevel = el.String

	systemPrompt := auraSystemPrompt
	if currentUser.RiskTolerance != "" {
		systemPrompt += "\nUser Risk: " + currentUser.RiskTolerance
	}

	contents := make([]models.GContent, 0, len(req.History)+1)
	for _, turn := range req.History {
		role := "user"
		if turn.Role == "aura" {
			role = "model"
		}
		contents = append(contents, models.GContent{Role: role, Parts: []models.GPart{{Text: turn.Text}}})
	}
	contents = append(contents, models.GContent{Role: "user", Parts: []models.GPart{{Text: req.Message}}})

	model := req.Model
	if model == "" {
		model = models.DefaultModel
	}

	text, strategy, err := runAuraAgent(apiKey, model, systemPrompt, contents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Text: text, Strategy: strategy})
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
			return "", nil, fmt.Errorf("empty response")
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
	return "", nil, fmt.Errorf("limit exceeded")
}

func executeTool(fc models.GFuncCall) (map[string]interface{}, *models.StrategyCardData) {
	switch fc.Name {
	case "get_stock_quote":
		ticker, _ := fc.Args["ticker"].(string)
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
		return map[string]interface{}{"status": "displayed"}, setup
	}
	return map[string]interface{}{"error": "unknown"}, nil
}

func argString(args map[string]interface{}, key string) string { v, _ := args[key].(string); return v }
func argFloat(args map[string]interface{}, key string) float64 {
	switch v := args[key].(type) {
	case float64: return v
	case int: return float64(v)
	}
	return 0
}

func callGeminiChat(apiKey, model, systemPrompt string, contents []models.GContent) (*models.GAPIResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	payload := map[string]interface{}{
		"system_instruction": map[string]interface{}{"parts": []map[string]interface{}{{"text": systemPrompt}}},
		"contents":           contents,
		"tools":              auraTools,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResp models.GAPIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	return &apiResp, nil
}
