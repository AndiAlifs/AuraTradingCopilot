package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"aura-trade/pkg/models"
)

// ollamaBaseURL is the Ollama local server address.
func ollamaBaseURL() string {
	if u := os.Getenv("OLLAMA_BASE_URL"); u != "" {
		u = strings.TrimSuffix(u, "/")
		u = strings.TrimSuffix(u, "/api")
		return u
	}
	return "http://localhost:11434"
}

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

// CallOllamaChat sends a chat turn to a locally-running Ollama model.
func CallOllamaChat(model, systemPrompt string, contents []models.GContent) (string, error) {
	messages := make([]ollamaMessage, 0, len(contents)+1)

	if systemPrompt != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: systemPrompt})
	}

	for _, c := range contents {
		role := c.Role
		if role == "model" {
			role = "assistant"
		}
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
		return "", fmt.Errorf("failed to reach Ollama: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response")
	}
	if ollamaResp.Error != "" {
		return "", fmt.Errorf("Ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Message.Content, nil
}
