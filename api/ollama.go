package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// ollamaBaseURL is the Ollama local server address.
// Override with the OLLAMA_BASE_URL environment variable if needed.
func ollamaBaseURL() string {
	if u := os.Getenv("OLLAMA_BASE_URL"); u != "" {
		return u
	}
	return "http://localhost:11434"
}

// ── Ollama wire types (https://docs.ollama.com/api/introduction) ──────────────

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
	Error   string        `json:"error,omitempty"`
}

// callOllamaChat sends a chat turn to a locally-running Ollama model.
// Because Ollama models generally don't support structured function-calling,
// this path is a single-shot request: system prompt + history + user message
// → assistant text response only (no tool execution).
func callOllamaChat(model, systemPrompt string, contents []gContent) (string, error) {
	messages := make([]ollamaMessage, 0, len(contents)+1)

	// Prepend system prompt
	if systemPrompt != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: systemPrompt})
	}

	// Convert Gemini-style gContent history to Ollama messages
	for _, c := range contents {
		role := c.Role
		if role == "model" {
			role = "assistant"
		}
		// Collapse all text parts into one message string
		text := ""
		for _, p := range c.Parts {
			if p.Text != "" {
				if text != "" {
					text += "\n"
				}
				text += p.Text
			}
		}
		if text != "" {
			messages = append(messages, ollamaMessage{Role: role, Content: text})
		}
	}

	payload := ollamaChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := ollamaBaseURL() + "/api/chat"
	log.Printf("[ollama/chat] → POST model=%s turns=%d", model, len(messages))

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("OLLAMA_API_KEY"); key != "" {
		httpReq.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Printf("[ollama/chat] ✗ network error: %v", err)
		return "", fmt.Errorf("failed to reach Ollama at %s — is it running? (%v)", ollamaBaseURL(), err)
	}
	defer resp.Body.Close()
	log.Printf("[ollama/chat] ← %d", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		log.Printf("[ollama/chat] ✗ parse error, raw body: %s", string(respBody))
		return "", fmt.Errorf("failed to parse Ollama response: %s", string(respBody))
	}
	if ollamaResp.Error != "" {
		log.Printf("[ollama/chat] ✗ API error: %s", ollamaResp.Error)
		return "", fmt.Errorf("Ollama error: %s", ollamaResp.Error)
	}

	log.Printf("[ollama/chat] done, content length=%d", len(ollamaResp.Message.Content))
	return ollamaResp.Message.Content, nil
}
