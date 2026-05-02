package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbInstance *sql.DB
	dbOnce     sync.Once
)

// User represents a trading platform user with psychological profile fields.
type User struct {
	ID                int    `json:"id"`
	Email             string `json:"email"`
	HashedPassword    string `json:"-"`
	RiskTolerance     string `json:"riskTolerance"`
	PreferredStrategy string `json:"preferredStrategy"`
	ExperienceLevel   string `json:"experienceLevel"`
	CreatedAt         string `json:"createdAt"`
}

// GetDB returns the singleton database connection, initializing it on first call.
func GetDB() *sql.DB {
	dbOnce.Do(func() {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbname)

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("[db] failed to open database: %v", err)
		}

		if err := db.Ping(); err != nil {
			log.Fatalf("[db] failed to connect to database: %v", err)
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)

		if err := migrateDB(db); err != nil {
			log.Fatalf("[db] migration failed: %v", err)
		}
		dbInstance = db
		log.Printf("[db] connected to MySQL database %s at %s:%s", dbname, host, port)
	})
	return dbInstance
}

func migrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id                 INT AUTO_INCREMENT PRIMARY KEY,
			email              VARCHAR(255) UNIQUE NOT NULL,
			hashed_password    TEXT NOT NULL,
			risk_tolerance     VARCHAR(50) DEFAULT '',
			preferred_strategy VARCHAR(50) DEFAULT '',
			experience_level   VARCHAR(50) DEFAULT '',
			created_at         DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	// Quick migration snippet to alter existing table (handles SQLite/MySQL column additions)
	// Ignore errors here since columns might already exist
	db.Exec(`ALTER TABLE users ADD COLUMN risk_tolerance VARCHAR(50) DEFAULT '';`)
	db.Exec(`ALTER TABLE users ADD COLUMN preferred_strategy VARCHAR(50) DEFAULT '';`)
	db.Exec(`ALTER TABLE users ADD COLUMN experience_level VARCHAR(50) DEFAULT '';`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_alerts (
			id           INT AUTO_INCREMENT PRIMARY KEY,
			user_id      INT NOT NULL,
			ticker       VARCHAR(255) NOT NULL,
			` + "`condition`" + `  VARCHAR(50) NOT NULL,
			target_price REAL NOT NULL,
			is_active    TINYINT DEFAULT 1,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_preferences (
			id                   INT AUTO_INCREMENT PRIMARY KEY,
			user_id              INT UNIQUE NOT NULL,
			default_exchange     VARCHAR(50) DEFAULT 'JK',
			notification_enabled TINYINT DEFAULT 1,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	return err
}
