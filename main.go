package main

import (
	"fmt"
	"log"
	"net/http"

	"finalsprint/pkg/api"
	"finalsprint/pkg/db"
	"finalsprint/server"
)

func main() {

	dbFile := "scheduler.db"

	if err := db.Init(dbFile); err != nil {
		log.Fatalf("db initialization failed: %v", err)
	}
	defer db.Close()

	api.Init()

	port := server.GetPort()

	http.Handle("/", http.FileServer(http.Dir(server.WebDir)))

	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server cant start: %v", err)
	}
}
