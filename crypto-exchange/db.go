package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/gommon/log"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// initDB if testMode = true, everytime startup the app, it will rebuild database with schema and prepare mock data.
func initDB(testMode bool) (*sql.DB, error) {
	// WAL + synchronous=NORMAL removes the per-commit fsync that otherwise caps
	// end-to-end order throughput (the in-memory engine does ~1M/s; the fsync-per-commit
	// default was the bottleneck). busy_timeout lets writers wait on the WAL write-lock
	// instead of erroring under concurrency.
	// WAL + synchronous=NORMAL drops the per-commit fsync that otherwise capped
	// end-to-end throughput at ~60 orders/sec (the in-memory engine does ~1M/s;
	// fsync-per-commit was the bottleneck). This lifts it ~10x to ~600+/s.
	// NOTE: SQLite is a single-writer DB, so a connection *pool* only adds write-lock
	// contention (measured slower); one connection serializes writes cleanly and wins.
	dsn := "file:/app/exg.db?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Infof("Database initialized successfully")

	// Run SQL files on startup if testMode
	if testMode {
		if err := runSQLFilesWithTransaction(db); err != nil {
			return nil, fmt.Errorf("failed to run SQL files: %w", err)
		}
		log.Infof("DB schema and testing data initialized successfully")
	}

	return db, err
}
