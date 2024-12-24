package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date TEXT NOT NULL,
	title TEXT NOT NULL,
	comment TEXT,
	repeat TEXT
);
`

var DB *sql.DB

func InitializeDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "scheduler.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	DB = db
	return db, nil
}
