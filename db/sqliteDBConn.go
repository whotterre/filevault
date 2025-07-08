package db

import (
	"database/sql"
)

func GetSQLiteDBConn() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", "filevault.db")
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, email TEXT UNIQUE, password TEXT)")
	if err != nil {
		return nil, err
	}

	return conn, nil
}