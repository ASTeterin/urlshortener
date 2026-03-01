package main

import (
	"net/http"

	"github.com/ASTeterin/urlshortener/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.GetShortURL)
	mux.HandleFunc(`/{id}`, handler.GetURL)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
