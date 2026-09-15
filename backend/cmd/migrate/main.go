// Command migrate opens the SQLite database and applies outstanding migrations,
// then exits. Used by build.py option 13 (Initialize/migrate database).
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"timeline-intelligence/backend/internal/repository/sqlite"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	_ = godotenv.Load(".env")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/timeline.db"
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Printf("Database ready: %s\n", dbPath)
}
