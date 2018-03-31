package db

import (
	"github.com/jackc/pgx"
)

// Database is the api connection to the database
type Database struct {
	Conn *pgx.Conn
}

// Options is the postgres connection config
type Options struct {
	Host     string
	User     string
	Password string
	Database string
}

// Connect to postgres database with options
func Connect(opts Options) (*Database, error) {
	var db Database
	var err error

	db.Conn, err = pgx.Connect(pgx.ConnConfig{
		Host:     opts.Host,
		User:     opts.User,
		Password: opts.Password,
		Database: opts.Database,
	})

	if err != nil {
		return nil, err
	}

	return &db, nil
}
