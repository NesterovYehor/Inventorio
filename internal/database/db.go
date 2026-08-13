package database

import (
	"database/sql"
	"fmt"
	 _ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func Connect(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}
return &DB{conn: conn}, nil
}

func (db *DB) Close() error{
	return db.conn.Close()
}
