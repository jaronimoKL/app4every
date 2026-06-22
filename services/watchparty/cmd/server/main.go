package main

import (
	"log"

	"app4every/services/watchparty/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("Critical error running application: %v", err)
	}
}
