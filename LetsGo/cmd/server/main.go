package main

import (
	"log"

	"github.com/GergesHany/Event-Streaming-System/LetsGo/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Println("Server running...")
	log.Fatal(srv.ListenAndServe())
}
