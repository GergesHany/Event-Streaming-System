package main

import (
	"github.com/GergesHany/Event-Streaming-System/LetsGo/internal/server"
	"log"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Println("Server running...")
	log.Fatal(srv.ListenAndServe())
}
