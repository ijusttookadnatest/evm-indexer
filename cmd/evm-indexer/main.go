package main

import (
	// "context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github/ijusttookadnatest/indexer-evm/config"
	"github/ijusttookadnatest/indexer-evm/db"
	"github/ijusttookadnatest/indexer-evm/listener"
	"github/ijusttookadnatest/indexer-evm/processor"
	"github/ijusttookadnatest/indexer-evm/redis"

	"github.com/joho/godotenv"
)

func loadEnv(path string) (map[string]string, error) {
	envFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer envFile.Close()
	env, err := godotenv.Parse(envFile)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func main() {
	// ctx := context.Background()
	// env, err := loadEnv(".env")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err = config.Init(env) ; err != nil {
	// 	log.Fatalf("Error init configuration: %v", err)
	// }
	// if _, err = db.Get() ; err != nil {
	// 	log.Fatalf("Error connection to database : %v", err)
	// }
	// if err = db.ApplyPendingMigrations() ; err != nil {
	// 	log.Fatalf("Error applying pending migrations: %v", err)
	// }
	// if _, err = redis.Get() ; err != nil {
	// 	log.Fatalf("Error connection to cache : %v", err)
	// }
	// if err = listener.Backfill() ; err != nil {
	// 	log.Fatalf("Error backfilling: %v", err)
	// }
	// if err = processor.Processor() ; err != nil {
	// 	log.Fatalf("Error processing: %v", err)
	// }
	// fmt.Println("Starting server")
	// http.ListenAndServe("127.0.0.1:8080", nil)
	cfg := GetConfig()
	db, err := ConnectDatabase(cfg.URN)
	if err != nil {
		return
	}
	repo := NewBlockRepo(db)
	service := NewBlocService(repo)
	server := NewServer(cfg.ListenAddr, service)

	if err := server.server.ListenAndServe(); err != nil {
	 log.Fatalf("Server error: %v", err)
	}
}