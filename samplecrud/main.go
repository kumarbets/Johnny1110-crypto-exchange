package main

import (
	"log"
	"os"
)

func main() {
	// Allow tests/deploys to override the DB path and port via env vars.
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "products.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Open (and auto-migrate) the SQLite database.
	db, err := OpenDB(dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Build the router with the real database injected.
	router := NewRouter(db)

	log.Printf("listening on http://localhost:%s", port)

	// Start the HTTP server; Run blocks until the process is killed.
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
