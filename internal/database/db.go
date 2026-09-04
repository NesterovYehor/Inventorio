package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

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
	db := &DB{conn: conn}
	if err := db.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("Failed to create schema: %w", err)
	}
	return db, nil
}

func (db *DB) initSchema(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, schemaSQL); err != nil {
		return err
	}
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
