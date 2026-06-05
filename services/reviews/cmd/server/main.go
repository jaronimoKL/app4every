package main

import (
	"log"
	"app4every/services/reviews/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("reviews service error: %v", err)
	}
}
