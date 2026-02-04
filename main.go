package main

import (
	"log"
)

func main() {
	env, err := loadEnv(".env")
	if err != nil {
		log.Fatal(err)
	}
	initConfig(env)
	if _, err = GetDB() ; err != nil {
		log.Fatalf("Error connection to database : %v", err)
	}
	if err = applyPendingMigrations() ; err != nil {
		log.Fatalf("Error applying pending migrations: %v", err)
	}
	if _, err = GetCache() ; err != nil {
		log.Fatalf("Error connection to cache : %v", err)
	}
}