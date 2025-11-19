package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := getPort()

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server cant start: %v", err)
	}
}
