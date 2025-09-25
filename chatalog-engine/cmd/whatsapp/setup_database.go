package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func setupSQLiteDatabase(sqlitePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return nil, err
	}

	if err := CreateTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

// CreateTables creates the tables for Merchant and Product models in the SQLite database.
func CreateTables(db *sql.DB) error {
	merchantTable := `CREATE TABLE IF NOT EXISTS merchants (
		id TEXT PRIMARY KEY,
		name TEXT,
		phone TEXT
	);`

	productTable := `CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		merchant_id TEXT,
		name TEXT,
		price REAL,
		FOREIGN KEY (merchant_id) REFERENCES merchants(id)
	);`

	if _, err := db.Exec(merchantTable); err != nil {
		return err
	}
	if _, err := db.Exec(productTable); err != nil {
		return err
	}
	return nil
}
