package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"aura-trade/internal/db"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	loadDotEnv()

	// Initialize DB
	database := db.GetDB()
	if database == nil {
		log.Fatal("failed to initialize database")
	}

	// Seed users
	seedData(database)
	fmt.Println("Seeding complete!")
}

func seedData(db *sql.DB) {
	users := []struct {
		Email    string
		Password string
	}{
		{"demo@example.com", "password123"},
		{"trader@example.com", "trader123"},
		{"admin@auratrading.com", "admin1234"},
	}

	for _, u := range users {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", u.Email, err)
			continue
		}

		res, err := db.Exec(
			"INSERT INTO users (email, hashed_password) VALUES (?, ?)",
			u.Email, string(hashed),
		)

		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				fmt.Printf("User %s already exists. Skipping.\n", u.Email)
			} else {
				log.Printf("Failed to insert user %s: %v", u.Email, err)
			}
			continue
		}

		userID, err := res.LastInsertId()
		if err != nil {
			log.Printf("Failed to get last insert ID for user %s: %v", u.Email, err)
			continue
		}

		// Seed user preferences
		_, err = db.Exec("INSERT INTO user_preferences (user_id) VALUES (?)", userID)
		if err != nil {
			log.Printf("Failed to insert preferences for user %s: %v", u.Email, err)
		}

		// Seed some sample alerts
		_, err = db.Exec(`
			INSERT INTO user_alerts (user_id, ticker, `+"`condition`"+`, target_price, is_active)
			VALUES 
				(?, 'AAPL', 'below', 150.00, 1),
				(?, 'MSFT', 'above', 420.00, 1)
		`, userID, userID)
		if err != nil {
			log.Printf("Failed to insert alerts for user %s: %v", u.Email, err)
		}

		fmt.Printf("Successfully seeded user: %s (password: %s)\n", u.Email, u.Password)
	}
}

// loadDotEnv reads the project-root .env (two levels above this file's package).
func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
