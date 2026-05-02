package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	logFile *os.File
	mu      sync.Mutex
	once    sync.Once
)

type LogEntry struct {
	Time            string `json:"time"`
	Level           string `json:"level"`
	Handler         string `json:"handler"`
	User            string `json:"user,omitempty"`
	Model           string `json:"model,omitempty"`
	Message         string `json:"message,omitempty"`
	Ticker          string `json:"ticker,omitempty"`
	ResponseSnippet string `json:"response_snippet,omitempty"`
	Error           string `json:"error,omitempty"`
	DurationMs      int64  `json:"duration_ms,omitempty"`
}

func init() {
	once.Do(func() {
		f, err := os.OpenFile("aura.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[logger] failed to open aura.log: %v\n", err)
			return
		}
		logFile = f
	})
}

func write(entry LogEntry) {
	entry.Time = time.Now().Format(time.RFC3339)
	b, err := json.Marshal(entry)
	if err != nil {
		return
	}
	line := string(b) + "\n"

	mu.Lock()
	defer mu.Unlock()

	// Always print to stdout so it shows in the terminal
	fmt.Print(line)

	// Also write to file
	if logFile != nil {
		logFile.WriteString(line)
	}
}

// Info logs a general informational entry.
func Info(handler string, entry LogEntry) {
	entry.Level = "INFO"
	entry.Handler = handler
	write(entry)
}

// Error logs an error entry.
func Error(handler string, entry LogEntry) {
	entry.Level = "ERROR"
	entry.Handler = handler
	write(entry)
}

// snippet truncates a string to at most n runes, appending "…" if cut.
func Snippet(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
