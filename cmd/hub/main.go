package main

import (
	"log"
	"net/http"
	"os"

	"infra-platform/hub/server"
)

func main() {
	addr := os.Getenv("HUB_ADDR")
	if addr == "" {
		addr = ":8081"
	}
	srv := server.New()
	log.Printf("hub listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
