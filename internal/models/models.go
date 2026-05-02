package models

import "strings"

// ModelInfo describes an available AI model for the frontend dropdown.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"` // "google" or "ollama"
}

// AvailableModels is the authoritative list of selectable models.
var AvailableModels = []ModelInfo{
	// ── Google Gemini ────────────────────────────────────────────────────────
	{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "google"},
	{ID: "gemini-3.1-pro-preview", Name: "Gemini 3.1 Pro", Provider: "google"},
	{ID: "gemini-3-flash-preview", Name: "Gemini 3 Flash", Provider: "google"},
	{ID: "gemini-3.1-flash-lite-preview", Name: "Gemini 3.1 Flash Lite", Provider: "google"},

	// ── Ollama (local) ───────────────────────────────────────────────────────
	{ID: "kimi-k2.6", Name: "Kimi K2.6", Provider: "ollama"},
	{ID: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", Provider: "ollama"},
	{ID: "gemma4", Name: "Gemma 4", Provider: "ollama"},
	{ID: "glm-5.1", Name: "GLM 5.1", Provider: "ollama"},
}

// DefaultModel is the model used when none is specified by the client.
const DefaultModel = "gemini-2.5-pro"

// StrategyCardData is the shared output model for both the analyze endpoint
// and the chat agent's generate_trade_setup tool.
type StrategyCardData struct {
	Ticker     string  `json:"ticker"`
	Confidence int     `json:"confidence"`
	Entry      float64 `json:"entry"`
	TakeProfit float64 `json:"takeProfit"`
	StopLoss   float64 `json:"stopLoss"`
	Rationale  string  `json:"rationale"`

	// Technical Analysis
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

// ── Gemini API wire types ─────────────────────────────────────────────────────

type GContent struct {
	Role  string  `json:"role"`
	Parts []GPart `json:"parts"`
}

type GPart struct {
	Text             string     `json:"text,omitempty"`
	ThoughtSignature string     `json:"thoughtSignature,omitempty"`
	FunctionCall     *GFuncCall `json:"functionCall,omitempty"`
	FunctionResponse *GFuncResp `json:"functionResponse,omitempty"`
}

type GFuncCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type GFuncResp struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type GAPIResponse struct {
	Candidates []struct {
		Content      GContent `json:"content"`
		FinishReason string   `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ProviderOf returns "google" or "ollama" for a given model ID.
func ProviderOf(modelID string) string {
	for _, m := range AvailableModels {
		if m.ID == modelID {
			return m.Provider
		}
	}
	if strings.Contains(modelID, "gemini") {
		return "google"
	}
	return "google" // Fallback
}
