package main

import (
	"log"

	"app4every/services/notebook/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("notebook service error: %v", err)
	}
}
