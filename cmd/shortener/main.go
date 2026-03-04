package main

import (
	"log"

	"github.com/ASTeterin/urlshortener/internal/handler"
	"github.com/ASTeterin/urlshortener/internal/repository"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	repo := repository.NewUrlRepository()
	shortenerService := service.NewShortenerService(repo)
	h := handler.NewHandler(shortenerService)

	r := gin.Default()

	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(c.Writer, c.Request)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(c.Writer, c.Request)
	})

	if err := r.Run(`:8080`); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
