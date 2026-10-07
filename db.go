package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("sqlite", "hermes.db")
	if err != nil {
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS users (
		fingerprint TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL
	);
	`
	_, err = db.Exec(query)
	return err
}

func GetUsername(fingerprint string) (string, error) {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE fingerprint = ?", fingerprint).Scan(&username)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return username, err
}

func RegisterUser(fingerprint, username string) error {
	_, err := db.Exec("INSERT INTO users (fingerprint, username) VALUES (?, ?)", fingerprint, username)
	return err
}
