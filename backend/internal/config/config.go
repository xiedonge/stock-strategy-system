package config

import "os"

// Config holds server and storage settings derived from environment variables.
type Config struct {
	Port   string
	DBPath string
}

// Load reads environment variables and provides sensible defaults.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/stock.db"
	}

	return Config{Port: port, DBPath: dbPath}
}
