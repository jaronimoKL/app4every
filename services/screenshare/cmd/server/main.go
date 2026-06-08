package main

import (
	"log"
	"app4every/services/screenshare/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("screenshare service error: %v", err)
	}
}
