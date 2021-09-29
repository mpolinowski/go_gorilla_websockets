package main

import (
	"go_gorilla_websocket/internal/handlers"
	"log"
	"net/http"
)

func main() {
	m := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting Webserver on Port 8080")

	_ = http.ListenAndServe(":8080", m)
}