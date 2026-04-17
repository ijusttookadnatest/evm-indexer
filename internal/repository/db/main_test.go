//go:build integration

package repository

import (
	"database/sql"
	"log"
	"os"
	"testing"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		log.Fatal("TEST_DATABASE_DSN is not set")
	}

	var err error
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	if err = testDB.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	if err = RunUpMigrations(testDB); err != nil {
		log.Fatalf("Failed to run up migrations: %v", err)
	}

	code := m.Run()

	if err = RunDownMigrations(testDB); err != nil {
		log.Fatalf("Failed to run down migrations: %v", err)
	}
	testDB.Close()

	os.Exit(code)
}

func truncateAll(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE events, transactions, blocks CASCADE")
	if err != nil {
		t.Fatalf("failted to truncate: %v", err)
	}
}

func seedFixtures(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile("testdata/fixtures.sql")
	if err != nil {
		t.Fatalf("failed to load fixtures files")
	}
	_, err = testDB.Exec(string(data))
	if err != nil {
		t.Fatalf("failed to seed fixtures")
	}
}
