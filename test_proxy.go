package main

import (
	"fmt"
	"net/http"
)

func main() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			for key, val := range via[0].Header {
				req.Header[key] = val
			}
			fmt.Println("Redirecting to:", req.URL.String())
			return nil
		},
	}
	req, _ := http.NewRequest("GET", "https://cache.libria.fun/videos/media/ts/9542/1/1080/aa675e5f3fe5b528517d812182344011_00000.ts", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X)")
	req.Header.Set("Referer", "https://anilibria.top/")
	
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println("Status:", resp.StatusCode)
}
