package main

import (
	"log"
)

func main() {
	// Open (and auto-migrate) a SQLite file in the project directory.
	db, err := OpenDB("products.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Build the router with the real database injected.
	router := NewRouter(db)

	log.Println("listening on http://localhost:8080")

	// Start the HTTP server; Run blocks until the process is killed.
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
