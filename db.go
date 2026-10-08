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

	queries := []string{`
		CREATE TABLE IF NOT EXISTS users (
			fingerprint TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL
		);
		`,
		`CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			original_name TEXT,
			owner_fingerprint TEXT,
			storage_path TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS file_access (
			file_id TEXT,
			username TEXT,
			PRIMARY KEY (file_id, username)
		);`,
	}

	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}
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

func GetAllUsers() ([]string, error) {
	rows, err := db.Query("SELECT username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func SaveFileRecord(id, originalName, ownerFingerprint, storagePath string) error {
	_, err := db.Exec(
		"INSERT INTO files (id, original_name, owner_fingerprint, storage_path) VALUES (?, ?, ?, ?)",
		id, originalName, ownerFingerprint, storagePath,
	)
	return err
}

func GrantAccess(fileId, username string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO file_access (file_id, username) VALUES (?, ?)", fileId, username)
	return err
}
