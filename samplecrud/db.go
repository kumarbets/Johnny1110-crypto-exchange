package main

import (
	"database/sql"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (registers itself, no cgo)
)

// OpenDB opens a SQLite database at dsn and ensures the schema exists.
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// SQLite allows only ONE writer at a time. Capping the pool to a single
	// connection avoids "database is locked" errors under concurrent requests.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS products (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	name     TEXT    NOT NULL,
	price    REAL    NOT NULL DEFAULT 0,
	quantity INTEGER NOT NULL DEFAULT 0
);`
	_, err := db.Exec(schema)
	return err
}
