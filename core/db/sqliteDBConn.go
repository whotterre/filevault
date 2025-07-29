package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func GetSQLiteDBConn() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", "filevault.db")
	if err != nil {
		return nil, err
	}

	// Create users table
	_, err = conn.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            email TEXT UNIQUE,
            password TEXT
        );
    `)
	if err != nil {
		return nil, err
	}

	// Create files table
	_, err = conn.Exec(`
       CREATE TABLE IF NOT EXISTS files (
            id TEXT PRIMARY KEY,
            user_id TEXT,
            file_name TEXT,
            file_path TEXT,
            size TEXT,
			type TEXT,
            uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			parent_id TEXT,
			is_public INTEGER DEFAULT FALSE,
			FOREIGN KEY (user_id) REFERENCES users(id)
        );
    `)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
