package main

import (
	"database/sql"
	"fmt"
	"github.com/johnny1110/crypto-exchange/utils"
	"github.com/labstack/gommon/log"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// seedSystemCounters loads the current total orders/trades from the DB into the
// in-memory counters so the UI shows real historical totals from boot.
func seedSystemCounters(db *sql.DB) {
	var orders, trades int64
	if err := db.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&orders); err == nil {
		utils.SetOrdersPlaced(orders)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM trades`).Scan(&trades); err == nil {
		utils.SetTradesTotal(trades)
	}
	log.Infof("[Counters] seeded orders=%d trades=%d", orders, trades)
}

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

	// Ensure the 1-minute OHLCV table exists (added interval; identical shape to the
	// seeded ohlcv_15min table). Migration: an earlier build created this table without
	// the close_time column, so drop that empty one before (re)creating the correct shape.
	var hasCloseTime int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('ohlcv_1min') WHERE name='close_time'`).Scan(&hasCloseTime)
	if hasCloseTime == 0 {
		db.Exec(`DROP TABLE IF EXISTS ohlcv_1min`)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS ohlcv_1min (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		open_price REAL NOT NULL,
		high_price REAL NOT NULL,
		low_price REAL NOT NULL,
		close_price REAL NOT NULL,
		volume REAL NOT NULL DEFAULT 0,
		quote_volume REAL NOT NULL DEFAULT 0,
		open_time INTEGER NOT NULL,
		close_time INTEGER NOT NULL,
		trade_count INTEGER NOT NULL DEFAULT 0,
		is_closed INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(symbol, open_time)
	);
	CREATE INDEX IF NOT EXISTS idx_ohlcv_1min_symbol_time ON ohlcv_1min(symbol, open_time);`); err != nil {
		return nil, fmt.Errorf("failed to ensure ohlcv_1min table: %w", err)
	}

	// Seed the system-wide orders/trades counters from the DB so the UI shows the
	// true historical totals (they then grow live as new orders/trades happen).
	seedSystemCounters(db)

	// Run SQL files on startup if testMode
	if testMode {
		if err := runSQLFilesWithTransaction(db); err != nil {
			return nil, fmt.Errorf("failed to run SQL files: %w", err)
		}
		log.Infof("DB schema and testing data initialized successfully")
	}

	return db, err
}
