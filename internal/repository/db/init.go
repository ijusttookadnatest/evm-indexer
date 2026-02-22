package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"github/ijusttookadnatest/indexer-evm/internal/config"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func New(cfg config.Config) (*sql.DB,error) {
	var err error

	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8) // if too much : use too much ressources / too less : bottlethrot
	db.SetMaxIdleConns(3) // if too much : use too much ressources / too less : overhead of creation destruction connection 
	db.SetConnMaxLifetime(5 * time.Minute) // recycle co to avoid fw to timeout (connection reset error: zombie co)
	db.SetConnMaxIdleTime(2 * time.Minute) // to avoid too much ressources busy if no request

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Connection to database successfull")
	return db, nil
}

func RunUpMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
        return err
    }
    if err := goose.Up(db, "migrations"); err != nil {
        return err
    }
	return nil
}

func RunDownMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
        return err
    }
    if err := goose.Up(db, "migrations"); err != nil {
        return err
    }
	return nil
}