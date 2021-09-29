package main

import (
	"go_gorilla_websocket/internal/handlers"
	"net/http"

	"github.com/bmizerany/pat"
)

func routes() http.Handler {
	m := pat.New()

	m.Get("/", http.HandlerFunc(handlers.Home))
	m.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	fileServer := http.FileServer(http.Dir("./static/"))
	m.Get("/static/", http.StripPrefix("/static", fileServer))

	return m
}