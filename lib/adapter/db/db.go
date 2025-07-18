package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type DB struct {
	conn *sql.DB
}

func Connect(cfg Config) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	var err error
	db := &DB{}
	db.conn, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.conn.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Shutdown(ctx context.Context) error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}
