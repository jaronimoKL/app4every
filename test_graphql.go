package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	query := `query {
		userRates(userId: 1, limit: 50, page: %d) {
			id
		}
	}`

	for page := 1; page <= 5; page++ {
		q := fmt.Sprintf(query, page)
		payload := map[string]string{"query": q}
		b, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "https://shikimori.io/api/graphql", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "App4Every")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		var out struct {
			Data struct {
				UserRates []struct {
					ID string `json:"id"`
				} `json:"userRates"`
			} `json:"data"`
		}
		
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &out)
		fmt.Printf("Page %d: %d items\n", page, len(out.Data.UserRates))
	}
}
