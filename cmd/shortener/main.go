package main

import (
	"net/http"

	"github.com/ASTeterin/urlshortener/internal/handler"
	"github.com/ASTeterin/urlshortener/internal/repository"
	"github.com/ASTeterin/urlshortener/internal/service"
)

func main() {
	repo := repository.NewUrlRepository()
	shortenerService := service.NewShortenerService(repo)
	h := handler.NewHandler(shortenerService)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, h.GetShortURL)
	mux.HandleFunc(`/{id}`, h.GetURL)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
