package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/app4every?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	var count int
	err = pool.QueryRow(context.Background(), "SELECT count(*) FROM reviews").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	fmt.Printf("Total reviews in DB: %d\n", count)
}
