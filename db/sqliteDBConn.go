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
            id INTEGER PRIMARY KEY,
            email TEXT UNIQUE,
            password TEXT,
            file_id INTEGER REFERENCES files(id) ON DELETE CASCADE
        );
    `)
    if err != nil {
        return nil, err
    }

    // Create files table
    _, err = conn.Exec(`
        CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY,
            user_id INTEGER,
            file_name TEXT,
            file_path TEXT,
            size TEXT,
            uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)
    if err != nil {
        return nil, err
    }

    return conn, nil
}