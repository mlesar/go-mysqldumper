package mysqldumper

import (
	"database/sql"
)

type DBWriter struct {
	db *sql.DB
}

func (s *DBWriter) Write(data string) error {
	_, err := s.db.Exec(data)
	return err
}

func NewDBWriter(db *sql.DB) *DBWriter {
	return &DBWriter{db: db}
}
