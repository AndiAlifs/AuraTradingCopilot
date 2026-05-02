package api

import (
	"encoding/json"
	"net/http"
)

// ModelInfo describes an available AI model for the frontend dropdown.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"` // "google" or "ollama"
}

// AvailableModels is the authoritative list of selectable models.
var AvailableModels = []ModelInfo{
	// ── Google Gemini ────────────────────────────────────────────────────────
	{ID: "gemini-2.5-pro",                 Name: "Gemini 2.5 Pro",               Provider: "google"},
	{ID: "gemini-3.1-pro-preview",         Name: "Gemini 3.1 Pro",               Provider: "google"},
	{ID: "gemini-3-flash-preview",         Name: "Gemini 3 Flash",               Provider: "google"},
	{ID: "gemini-3.1-flash-lite-preview",  Name: "Gemini 3.1 Flash Lite",        Provider: "google"},

	// ── Ollama (local) ───────────────────────────────────────────────────────
	{ID: "kimi-k2.6",       Name: "Kimi K2.6",       Provider: "ollama"},
	{ID: "deepseek-v4-pro", Name: "DeepSeek V4 Pro",  Provider: "ollama"},
	{ID: "gemma4",          Name: "Gemma 4",           Provider: "ollama"},
	{ID: "glm-5.1",         Name: "GLM 5.1",           Provider: "ollama"},
}

// DefaultModel is the model used when none is specified by the client.
const DefaultModel = "gemini-2.5-pro"

// ModelsHandler serves GET /api/models — no auth required so the UI can
// populate the dropdown before login if needed.
func ModelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AvailableModels)
}

// providerOf returns "google" or "ollama" for a given model ID.
// Falls back to "google" for unrecognised models.
func providerOf(modelID string) string {
	for _, m := range AvailableModels {
		if m.ID == modelID {
			return m.Provider
		}
	}
	return "google"
}
