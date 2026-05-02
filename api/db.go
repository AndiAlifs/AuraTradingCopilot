package api

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	dbInstance *sql.DB
	dbOnce     sync.Once
)

// GetDB returns the singleton database connection, initializing it on first call.
func GetDB() *sql.DB {
	dbOnce.Do(func() {
		path := os.Getenv("DATABASE_PATH")
		if path == "" {
			path = "./aura.db"
		}
		db, err := sql.Open("sqlite", path)
		if err != nil {
			log.Fatalf("[db] failed to open database: %v", err)
		}
		db.SetMaxOpenConns(1) // SQLite is single-writer
		if err := migrateDB(db); err != nil {
			log.Fatalf("[db] migration failed: %v", err)
		}
		dbInstance = db
		log.Printf("[db] initialized at %s", path)
	})
	return dbInstance
}

func migrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			email            TEXT UNIQUE NOT NULL,
			hashed_password  TEXT NOT NULL,
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS user_alerts (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id      INTEGER NOT NULL,
			ticker       TEXT NOT NULL,
			condition    TEXT NOT NULL,
			target_price REAL NOT NULL,
			is_active    INTEGER DEFAULT 1,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS user_preferences (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id              INTEGER UNIQUE NOT NULL,
			default_exchange     TEXT DEFAULT 'JK',
			notification_enabled INTEGER DEFAULT 1,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	return err
}
