package main

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("sqlite", "hermes.db")
	if err != nil {
		return err
	}

	// Make sure storage folder actually exists
	_ = os.MkdirAll("storage", 0755)

	// All the tables to make
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

	// loop through and execute each query to make the tables
	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}
	return err
}

// DATABASE FUNCTIONS, these are for interacting with db so i dont need to write more sql
// All the below are db funcs

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
		// The &u means it is a pointer so scan writes to u and not a copy of it
		// go being weird
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

type FileInfo struct {
	ID           string
	OriginalName string
	Owner        string
}

func GetAccessibleFiles(username string) ([]FileInfo, error) {
	query := `
		SELECT f.id, f.original_name, u.username
		FROM files f
		JOIN file_access fa ON f.id = fa.file_id
		JOIN users u ON f.owner_fingerprint = u.fingerprint
		WHERE fa.username = ?
	`
	rows, err := db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileInfo
	for rows.Next() {
		var file FileInfo
		if err := rows.Scan(&file.ID, &file.OriginalName, &file.Owner); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func GetFileForUser(fileId, username string) (storagePath string, err error) {
	query := `
		SELECT f.storage_path
		FROM files f
		JOIN file_access fa ON f.id = fa.file_id
		WHERE f.id = ? AND fa.username = ?
	`
	err = db.QueryRow(query, fileId, username).Scan(&storagePath)
	return storagePath, err
}

func GetUnsharedFileForUser(fingerprint string) (id string, filename string, err error) {
	query := `
		SELECT f.id, f.original_name
		FROM files f
		LEFT JOIN file_access fa ON f.id = fa.file_id
		WHERE f.owner_fingerprint = ? AND fa.file_id IS NULL
		LIMIT 1
	`
	err = db.QueryRow(query, fingerprint).Scan(&id, &filename)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return id, filename, err
}
