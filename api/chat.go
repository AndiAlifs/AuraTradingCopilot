package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// ── Public API types ──────────────────────────────────────────────────────────

type ChatRequest struct {
	History []ConversationTurn `json:"history"`
	Message string             `json:"message"`
}

type ConversationTurn struct {
	Role string `json:"role"` // "user" or "aura"
	Text string `json:"text"`
}

type ChatResponse struct {
	Text     string            `json:"text"`
	Strategy *StrategyCardData `json:"strategy,omitempty"`
}

// ── Gemini API wire types ─────────────────────────────────────────────────────

type gContent struct {
	Role  string  `json:"role"`
	Parts []gPart `json:"parts"`
}

type gPart struct {
	Text             string     `json:"text,omitempty"`
	ThoughtSignature string     `json:"thoughtSignature,omitempty"`
	FunctionCall     *gFuncCall `json:"functionCall,omitempty"`
	FunctionResponse *gFuncResp `json:"functionResponse,omitempty"`
}

type gFuncCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type gFuncResp struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type gAPIResponse struct {
	Candidates []struct {
		Content      gContent `json:"content"`
		FinishReason string   `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ── System prompt + tool declarations ────────────────────────────────────────

const auraSystemPrompt = `You are Aura, an AI quantitative analyst built for Indonesian Stock Exchange (IDX) swing trading.

Personality: Sharp, confident, trader-brained. Talk like someone who has been at the desk for years — direct, no hedging, no corporate-speak. Keep it tight.

Tools you have:
1. get_stock_quote — pulls live price, volume, and 5-day price history from Yahoo Finance
2. generate_trade_setup — renders a structured trade card (entry, target, stop) to the user

Behavior rules:
- User asks about a stock or wants a setup → ALWAYS call get_stock_quote first, then generate_trade_setup with levels derived from real data
- Use the recent price history to identify support/resistance for your levels
- Take profit: target 4–7% above current price for IDX swing trades
- Stop loss: 2–4% below entry, respect recent support
- Confidence: 50–95, based on trend clarity, volume, and momentum alignment
- For general questions (how you work, market theory, risk management) → respond conversationally, no tools needed
- NEVER invent prices — always fetch them first
- Conversational text: 1–3 punchy sentences max
- After generate_trade_setup, add one sentence about risk or what to watch`

var auraTools = []map[string]interface{}{
	{
		"functionDeclarations": []map[string]interface{}{
			{
				"name":        "get_stock_quote",
				"description": "Fetch real-time stock price, volume, and 5-day close history from Yahoo Finance for an IDX stock",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"ticker": map[string]interface{}{
							"type":        "string",
							"description": "IDX ticker with .JK suffix, e.g. BBCA.JK or GOTO.JK",
						},
					},
					"required": []string{"ticker"},
				},
			},
			{
				"name":        "generate_trade_setup",
				"description": "Display a structured trade setup card to the user with entry, take profit, stop loss and rationale",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"ticker":     map[string]interface{}{"type": "string", "description": "Stock ticker symbol"},
						"confidence": map[string]interface{}{"type": "number", "description": "Signal confidence 0–100"},
						"entry":      map[string]interface{}{"type": "number", "description": "Recommended entry price"},
						"takeProfit": map[string]interface{}{"type": "number", "description": "Take profit target price"},
						"stopLoss":   map[string]interface{}{"type": "number", "description": "Stop loss price"},
						"rationale":  map[string]interface{}{"type": "string", "description": "2 punchy sentences: what the setup is and why now"},
					},
					"required": []string{"ticker", "confidence", "entry", "takeProfit", "stopLoss", "rationale"},
				},
			},
		},
	},
}

// ── Handler ───────────────────────────────────────────────────────────────────

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[chat] %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Printf("[chat] ERROR: GEMINI_API_KEY is not set")
		http.Error(w, "GEMINI_API_KEY is not configured", http.StatusInternalServerError)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[chat] ERROR decoding request: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Build Gemini history from simple conversation turns
	contents := make([]gContent, 0, len(req.History)+1)
	for _, turn := range req.History {
		if turn.Text == "" {
			continue
		}
		role := "user"
		if turn.Role == "aura" {
			role = "model"
		}
		contents = append(contents, gContent{
			Role:  role,
			Parts: []gPart{{Text: turn.Text}},
		})
	}
	contents = append(contents, gContent{
		Role:  "user",
		Parts: []gPart{{Text: req.Message}},
	})

	text, strategy, err := runAuraAgent(apiKey, contents)
	if err != nil {
		log.Printf("[chat] ERROR from agent: %v", err)
		http.Error(w, "agent error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Text: text, Strategy: strategy})
}

// ── Agent loop ────────────────────────────────────────────────────────────────

func runAuraAgent(apiKey string, contents []gContent) (string, *StrategyCardData, error) {
	var strategy *StrategyCardData

	for i := 0; i < 8; i++ {
		resp, err := callGeminiChat(apiKey, contents)
		if err != nil {
			return "", nil, err
		}
		if len(resp.Candidates) == 0 {
			return "", nil, fmt.Errorf("empty response from Gemini")
		}

		candidate := resp.Candidates[0]
		var textParts []string
		var funcCalls []gFuncCall

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

		// Append the model's tool-call turn to history
		contents = append(contents, gContent{
			Role:  "model",
			Parts: candidate.Content.Parts,
		})

		// Execute tools and collect responses
		var responseParts []gPart
		for _, fc := range funcCalls {
			result, setupData := executeTool(fc)
			if setupData != nil {
				strategy = setupData
			}
			responseParts = append(responseParts, gPart{
				FunctionResponse: &gFuncResp{
					Name:     fc.Name,
					Response: result,
				},
			})
		}

		// Function responses use "user" role in Gemini API
		contents = append(contents, gContent{
			Role:  "user",
			Parts: responseParts,
		})
	}

	return "", nil, fmt.Errorf("agent loop exceeded iteration limit")
}

// ── Tool execution ────────────────────────────────────────────────────────────

func executeTool(fc gFuncCall) (map[string]interface{}, *StrategyCardData) {
	switch fc.Name {
	case "get_stock_quote":
		ticker, _ := fc.Args["ticker"].(string)
		if ticker == "" {
			return map[string]interface{}{"error": "ticker is required"}, nil
		}
		if !strings.HasSuffix(strings.ToUpper(ticker), ".JK") {
			ticker = strings.ToUpper(ticker) + ".JK"
		}
		quote, err := fetchYahooQuote(ticker)
		if err != nil {
			return map[string]interface{}{"error": err.Error()}, nil
		}
		return quote, nil

	case "generate_trade_setup":
		setup := &StrategyCardData{
			Ticker:     argString(fc.Args, "ticker"),
			Confidence: int(argFloat(fc.Args, "confidence")),
			Entry:      argFloat(fc.Args, "entry"),
			TakeProfit: argFloat(fc.Args, "takeProfit"),
			StopLoss:   argFloat(fc.Args, "stopLoss"),
			Rationale:  argString(fc.Args, "rationale"),
		}
		return map[string]interface{}{"status": "card displayed to user"}, setup
	}

	return map[string]interface{}{"error": "unknown tool: " + fc.Name}, nil
}

func argString(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

func argFloat(args map[string]interface{}, key string) float64 {
	switch v := args[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0
}

// ── Gemini API call ───────────────────────────────────────────────────────────

func callGeminiChat(apiKey string, contents []gContent) (*gAPIResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-pro-preview:generateContent?key=%s", apiKey)

	log.Printf("[gemini/chat] → POST model=gemini-3.1-pro-preview turns=%d", len(contents))
	if len(contents) > 0 {
		last := contents[len(contents)-1]
		if len(last.Parts) > 0 {
			if last.Parts[0].FunctionResponse != nil {
				log.Printf("[gemini/chat]   last %s: functionResponse name=%s", last.Role, last.Parts[0].FunctionResponse.Name)
			} else if last.Parts[0].FunctionCall != nil {
				log.Printf("[gemini/chat]   last %s: functionCall name=%s", last.Role, last.Parts[0].FunctionCall.Name)
			} else if t := last.Parts[0].Text; t != "" {
				preview := t
				if len(preview) > 120 {
					preview = preview[:120] + "..."
				}
				log.Printf("[gemini/chat]   last %s: %q", last.Role, preview)
			}
		}
	}

	payload := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]interface{}{{"text": auraSystemPrompt}},
		},
		"contents": contents,
		"tools":    auraTools,
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 1024,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[gemini/chat] ✗ network error: %v", err)
		return nil, fmt.Errorf("failed to call Gemini API: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("[gemini/chat] ← %d", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp gAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		log.Printf("[gemini/chat] ✗ parse error, raw body: %s", string(respBody))
		return nil, fmt.Errorf("failed to parse Gemini response: %s", string(respBody))
	}
	if apiResp.Error != nil {
		log.Printf("[gemini/chat] ✗ API error: %s", apiResp.Error.Message)
		return nil, fmt.Errorf("Gemini API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Candidates) > 0 {
		c := apiResp.Candidates[0]
		funcCalls := 0
		for _, p := range c.Content.Parts {
			if p.FunctionCall != nil {
				funcCalls++
				log.Printf("[gemini/chat]   funcCall: %s args=%v", p.FunctionCall.Name, p.FunctionCall.Args)
			}
		}
		log.Printf("[gemini/chat]   finish=%s parts=%d funcCalls=%d", c.FinishReason, len(c.Content.Parts), funcCalls)
	}

	return &apiResp, nil
}
