package aeroplane

import (
	"database/sql"
)

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func execSQL(path string, sqlText string) error {
	db, err := openDB(path)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(sqlText)
	return err
}
